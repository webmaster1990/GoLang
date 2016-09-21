package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"

	"database/sql"

	"strconv"

	"io"
	"os"

	"time"

	"github.com/gorilla/context"
	"golang.org/x/crypto/bcrypt"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

const USER = "user"

var userSessions *SessionStorage
var invalidApiKey = errors.New("invalid API key")

type Response struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusNotFound, Response{})
}

func index(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, Response{})
}

func JSON(w http.ResponseWriter, code int, obj interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(obj)
}

func verifyApiKey(r *http.Request) (string, error) {
	apiKey := r.Header.Get("X-Api-Key")
	if len(apiKey) == 0 {
		return "", invalidApiKey
	}
	id, ok := userSessions.Get(apiKey)
	if ok {
		return id, nil
	}
	return "", invalidApiKey
}

func login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	var userId string
	var hashedPassword []byte
	err := pg.QueryRow("SELECT user_id, password FROM users WHERE lower(email) = $1", email).Scan(&userId, &hashedPassword)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		JSON(w, http.StatusUnauthorized, Response{nil, "login failed"})
		return
	}

	apiKey := createApiKey(userId)
	userSessions.Set(apiKey, userId)
	JSON(w, http.StatusOK, Response{apiKey, "login success"})
}

func dbGetUser(userId string) (User, error) {
	var user User
	err := pg.QueryRow("SELECT organization_id, full_name, is_admin FROM users WHERE user_id = $1", userId).Scan(&user.OrganizationId, &user.FullName, &user.IsAdmin)
	if err != nil {
		return user, err
	}
	user.UserId = userId
	return user, err
}

func getProjects(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)

	rows, err := pg.Query(`
		SELECT
		  project_id, project_name, coalesce(logo_url, ''), description, budget, donor,
		  coalesce(mission, ''), coalesce(vision, ''), to_char(timeline_from, 'YYYY-MM-DD HH:MI:SS TZ'),
		  to_char(timeline_to, 'YYYY-MM-DD HH:MI:SS TZ'),
		  array_remove(array_agg(boundary_partner_id), NULL), array_remove(array_agg(partner_name), NULL),
		  array_remove(array_agg(resource_id), NULL), array_remove(array_agg(resource_url), NULL)
		FROM projects
		LEFT JOIN boundary_partners USING (project_id)
		LEFT JOIN external_resources USING (project_id)
		WHERE organization_id = $1
		GROUP BY project_id
	`, user.OrganizationId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	projects := []Project{}

	for rows.Next() {
		p := Project{}
		err = rows.Scan(&p.ProjectId, &p.ProjectName, &p.LogoUrl, &p.Description, &p.Budget, &p.Donor, &p.Mission, &p.Vision,
			&p.TimelineFrom, &p.TimelineTo, &p.BoundaryPartnerIds, &p.BoundaryPartnerNames, &p.ResourceIds, &p.ResourceUrls)
		if err != nil {
			rows.Close()
			JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
			return
		}
		projects = append(projects, p)
	}

	JSON(w, http.StatusOK, Response{projects, "success"})
}

func addProject(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)

	// only add project if current user is an admin
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	var input Project
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&input); err != nil {
		JSON(w, http.StatusBadRequest, Response{nil, err.Error()})
		return
	}

	_, err := pg.Exec("INSERT INTO projects (project_id, organization_id, project_name, description, budget, donor, vision, mission) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		uuid.NewV4().String(), user.OrganizationId, input.ProjectName, input.Description, input.Budget, input.Donor, input.Vision, input.Mission)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func dbGetProjectOrganization(projectId string) (string, error) {
	var organizationId string
	err := pg.QueryRow("SELECT organization_id FROM projects WHERE project_id = $1", projectId).Scan(&organizationId)
	if err != nil {
		return organizationId, err
	}
	return organizationId, err
}

func addBoundaryPartner(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)

	// only add boundary partner if current user is an admin
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	partnerName := r.FormValue("partner_name")

	_, err = pg.Exec("INSERT INTO boundary_partners (boundary_partner_id, project_id, partner_name) VALUES ($1, $2, $3)",
		uuid.NewV4().String(), projectId, partnerName)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}
	JSON(w, http.StatusOK, Response{nil, "success"})
}

func getBoundaryPartner(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	// brute force queries
	// TODO: optimize?
	var bp BoundaryPartner
	row := pg.QueryRow(`
		SELECT boundary_partner_id, project_id, partner_name, coalesce(outcome_statement, '')
		FROM boundary_partners
		JOIN projects USING (project_id)
		LEFT JOIN progress_markers USING (boundary_partner_id)
		WHERE organization_id = $1
		GROUP BY boundary_partner_id
	`, user.OrganizationId)

	err = row.Scan(&bp.BoundaryPartnerId, &bp.ProjectId, &bp.PartnerName, &bp.OutcomeStatement)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	// get progress markers for boundary partner
	var markers []*ProgressMarker
	rows, err := pg.Query("SELECT progress_marker_id, title, type, order_number FROM progress_markers WHERE boundary_partner_id = $1 ORDER BY ts_created", bp.BoundaryPartnerId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}
	for rows.Next() {
		pm := ProgressMarker{}
		pm.BoundaryPartnerId = bp.BoundaryPartnerId
		err = rows.Scan(&pm.ProgressMarkerId, &pm.Title, &pm.Type, &pm.OrderNumber)
		if err != nil {
			rows.Close()
			JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
			return
		}
		markers = append(markers, &pm)
	}
	rows.Close()

	// get challenges for each progress marker
	for i, pm := range markers {
		var challenges []*Challenge
		rows, err = pg.Query("SELECT challenge_id, challenge_name FROM challenges WHERE progress_marker_id = $1 ORDER BY ts_created", pm.ProgressMarkerId)
		if err != nil {
			JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
			return
		}
		for rows.Next() {
			c := Challenge{}
			c.ProgressMarkerId = pm.ProgressMarkerId
			err = rows.Scan(&c.ChallengeId, &c.ChallengeName)
			if err != nil {
				rows.Close()
				JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
				return
			}
			challenges = append(challenges, &c)
		}
		rows.Close()
		markers[i].Challenges = challenges
	}

	// get strategies for each progress marker
	for i, pm := range markers {
		var strategies []*Strategy
		rows, err = pg.Query("SELECT strategy_id, strategy_name FROM strategies WHERE progress_marker_id = $1 ORDER BY ts_created", pm.ProgressMarkerId)
		if err != nil {
			JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
			return
		}
		for rows.Next() {
			s := Strategy{}
			s.ProgressMarkerId = pm.ProgressMarkerId
			err = rows.Scan(&s.StrategyId, &s.StrategyName)
			if err != nil {
				rows.Close()
				JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
				return
			}
			strategies = append(strategies, &s)
		}
		rows.Close()
		markers[i].Strategies = strategies
	}

	bp.ProgressMarkers = markers
	JSON(w, http.StatusOK, Response{bp, "success"})
}

func getBoundaryPartner1(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	rows, err := pg.Queryx(`
		SELECT boundary_partner_id, project_id, partner_name, coalesce(outcome_statement, '') AS outcome_statement,
		  progress_marker_id, title, type, order_number, challenge_id, challenge_name, strategy_id, strategy_name
		FROM boundary_partners
		JOIN projects USING (project_id)
		LEFT JOIN progress_markers pm USING (boundary_partner_id)
		LEFT JOIN challenges ch USING (progress_marker_id)
		LEFT JOIN strategies strats USING (progress_marker_id)
		WHERE organization_id = $1
		ORDER BY pm.ts_created, ch.ts_created
	`, projectOrganizationId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	type ChallengeQuery struct {
		BoundaryPartnerId string         `db:"boundary_partner_id"`
		ProjectId         string         `db:"project_id"`
		PartnerName       string         `db:"partner_name"`
		OutcomeStatement  string         `db:"outcome_statement"`
		ProgressMarkerId  string         `db:"progress_marker_id"`
		Title             string         `db:"title"`
		Type              int            `db:"type"`
		OrderNumber       int            `db:"order_number"`
		ChallengeId       sql.NullString `db:"challenge_id"`
		ChallengeName     sql.NullString `db:"challenge_name"`
		StrategyId        sql.NullString `db:"strategy_id"`
		StrategyName      sql.NullString `db:"strategy_name"`
	}
	var bp *BoundaryPartner
	pmDict := make(map[string]*ProgressMarker)
	challengeDict := make(map[string]bool)
	strategyDict := make(map[string]bool)
	for rows.Next() {
		var cq ChallengeQuery
		err = rows.StructScan(&cq)
		if err != nil {
			rows.Close()
			JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
			return
		}
		// set boundary partner if it does not exist (only one)
		if bp == nil {
			bp = &BoundaryPartner{}
			bp.ProjectId = cq.ProjectId
			bp.BoundaryPartnerId = cq.BoundaryPartnerId
			bp.PartnerName = cq.PartnerName
			bp.OutcomeStatement = cq.OutcomeStatement
		}
		// add a progress marker to the dictionary if it hasn't been added
		if _, exists := pmDict[cq.ProgressMarkerId]; !exists {
			tmp := ProgressMarker{}
			tmp.ProgressMarkerId = cq.ProgressMarkerId
			tmp.BoundaryPartnerId = bp.BoundaryPartnerId
			tmp.Title = cq.Title
			tmp.OrderNumber = cq.OrderNumber
			tmp.Type = cq.Type
			pmDict[cq.ProgressMarkerId] = &tmp
		}
		// add the challenge to the dictionary if it hasn't been added and link it to the progress marker
		if _, exists := challengeDict[cq.ChallengeId.String]; cq.ChallengeId.Valid && !exists {
			ch := Challenge{}
			ch.ChallengeId = cq.ChallengeId.String
			ch.ProgressMarkerId = cq.ProgressMarkerId
			ch.ChallengeName = cq.ChallengeName.String
			challengeDict[cq.ChallengeId.String] = true
			pmDict[cq.ProgressMarkerId].Challenges = append(pmDict[cq.ProgressMarkerId].Challenges, &ch)
		}
		// add the strategy to the dictionary if it hasn't been added and link it to the progress marker
		if _, exists := strategyDict[cq.StrategyId.String]; cq.StrategyId.Valid && !exists {
			strat := Strategy{}
			strat.StrategyId = cq.StrategyId.String
			strat.ProgressMarkerId = cq.ProgressMarkerId
			strat.StrategyName = cq.StrategyName.String
			strategyDict[cq.StrategyId.String] = true
			pmDict[cq.ProgressMarkerId].Strategies = append(pmDict[cq.ProgressMarkerId].Strategies, &strat)
		}
	}
	for _, v := range pmDict {
		bp.ProgressMarkers = append(bp.ProgressMarkers, v)
	}

	JSON(w, http.StatusOK, Response{*bp, "success"})
}

func addProgressMarker(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)

	// only add progress marker if current user is an admin
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	partnerId := mux.Vars(r)["partnerId"]
	var input ProgressMarker
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&input); err != nil {
		JSON(w, http.StatusBadRequest, Response{nil, err.Error()})
		return
	}

	var orderNumber int
	err = pg.QueryRow("SELECT count(*) FROM progress_markers WHERE boundary_partner_id = $1", partnerId).Scan(&orderNumber)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	_, err = pg.Exec("INSERT INTO progress_markers (progress_marker_id, boundary_partner_id, title, type, order_number) VALUES ($1, $2, $3, $4, $5)",
		uuid.NewV4().String(), partnerId, &input.Title, &input.Type, orderNumber+1)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}
	JSON(w, http.StatusOK, Response{nil, "success"})
}

func addChallenge(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)

	// only add challenge if current user is an admin
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	markerId := mux.Vars(r)["progressMarkerId"]
	challenge := r.FormValue("challenge")

	_, err = pg.Exec("INSERT INTO challenges (challenge_id, progress_marker_id, challenge_name) VALUES ($1, $2, $3)",
		uuid.NewV4().String(), markerId, challenge)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}
	JSON(w, http.StatusOK, Response{nil, "success"})
}

func addStrategy(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	markerId := mux.Vars(r)["progressMarkerId"]
	strategy := r.FormValue("strategy")

	_, err = pg.Exec("INSERT INTO strategies (strategy_id, progress_marker_id, strategy_name) VALUES ($1, $2, $3)",
		uuid.NewV4().String(), markerId, strategy)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}
	JSON(w, http.StatusOK, Response{nil, "success"})
}

func authenticate(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := verifyApiKey(r)
		if err != nil {
			JSON(w, http.StatusUnauthorized, Response{nil, err.Error()})
			return
		}
		user, err := dbGetUser(userId)
		if err != nil {
			JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
			return
		}
		context.Set(r, USER, user)
		fn(w, r)
		context.Clear(r)
	}
}

func updateProjectName(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectName := r.FormValue("project_name")

	_, err = pg.Exec("UPDATE projects SET project_name = $1 WHERE project_id = $2", projectName, projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateProjectLogo(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	r.ParseMultipartForm(10 * int64(1e6)) // grab the multipart form
	if r.MultipartForm == nil {
		JSON(w, http.StatusBadRequest, Response{nil, "request must be multipart/form-data"})
		return
	}

	file, handler, err := r.FormFile("project_logo") // grab the filenames

	if err != nil {
		JSON(w, http.StatusBadRequest, Response{nil, err.Error()})
		return
	}
	defer file.Close()

	//TODO	url, err := uploadFile(file, handler.Filename, projectId)
	url := handler.Filename
	if err != nil {
		JSON(w, http.StatusBadRequest, Response{nil, err.Error()})
		return
	}

	_, err = pg.Exec("UPDATE projects SET logo_url = $1 WHERE project_id = $2", url, projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateProjectBudget(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	budget, err := strconv.ParseFloat(r.FormValue("project_budget"), 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, Response{nil, err.Error()})
		return
	}

	_, err = pg.Exec("UPDATE projects SET budget = $1 WHERE project_id = $2", budget, projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateProjectTimeline(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	var input Project
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&input); err != nil {
		JSON(w, http.StatusBadRequest, Response{nil, err.Error()})
		return
	}

	t_f, err1 := time.Parse(time.RFC3339, input.TimelineFrom)
	t_t, err2 := time.Parse(time.RFC3339, input.TimelineTo)

	if err1 != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err1.Error()})
		return
	}
	if err2 != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err2.Error()})
		return
	}

	_, err = pg.Exec("UPDATE projects SET timeline_from = $1, timeline_to = $2 WHERE project_id = $3", t_f, t_t, projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateProjectDescription(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	description := r.FormValue("project_description")

	_, err = pg.Exec("UPDATE projects SET description = $1 WHERE project_id = $2", description, projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateProjectDonor(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	donor := r.FormValue("project_donor")

	_, err = pg.Exec("UPDATE projects SET donor = $1 WHERE project_id = $2", donor, projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateProjectMission(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	mission := r.FormValue("project_mission")

	_, err = pg.Exec("UPDATE projects SET mission = $1 WHERE project_id = $2", mission, projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateProjectVision(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	vision := r.FormValue("project_vision")

	_, err = pg.Exec("UPDATE projects SET vision = $1 WHERE project_id = $2", vision, projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	_, err = pg.Exec("DELETE FROM projects WHERE project_id = $1", projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func resetProjectName(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	_, err = pg.Exec("UPDATE projects SET project_name = NULL WHERE project_id = $1", projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, "update failed"})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func resetProjectLogo(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}
	// TODO delete logo file in S3
	_, err = pg.Exec("UPDATE projects SET logo_url = NULL WHERE project_id = $1", projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func resetProjectDescription(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	_, err = pg.Exec("UPDATE projects SET description = NULL WHERE project_id = $1", projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func resetProjectBudget(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	_, err = pg.Exec("UPDATE projects SET budget = NULL WHERE project_id = $1", projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func resetProjectTimeline(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	_, err = pg.Exec("UPDATE projects SET timeline_from = NULL, timeline_to = NULL WHERE project_id = $1", projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func resetProjectDonor(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	_, err = pg.Exec("UPDATE projects SET donor = NULL WHERE project_id = $1", projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func resetProjectMission(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	_, err = pg.Exec("UPDATE projects SET mission = NULL WHERE project_id = $1", projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func resetProjectVision(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	_, err = pg.Exec("UPDATE projects SET vision = NULL WHERE project_id = $1", projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updatePartnerName(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	partnerId := mux.Vars(r)["partnerId"]
	partnerName := r.FormValue("partner_name")

	_, err = pg.Exec("UPDATE boundary_partners SET partner_name = $1 WHERE boundary_partner_id = $2", partnerName, partnerId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateOutcomeStatement(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	partnerId := mux.Vars(r)["partnerId"]
	partnerOutcomeStatement := r.FormValue("outcome_statement")

	_, err = pg.Exec("UPDATE boundary_partners SET outcome_statement = $1 WHERE boundary_partner_id = $2", partnerOutcomeStatement, partnerId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateProgressMarker(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	progressMarkerId := mux.Vars(r)["progressMarkerId"]
	var input ProgressMarker
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&input); err != nil {
		JSON(w, http.StatusBadRequest, Response{nil, err.Error()})
		return
	}

	var oldOrderNumber int
	var boundaryPartnerId string
	err = pg.QueryRow("SELECT order_number, boundary_partner_id from progress_markers where progress_marker_id = $1", progressMarkerId).Scan(&oldOrderNumber, &boundaryPartnerId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	if oldOrderNumber != input.OrderNumber {
		if oldOrderNumber > input.OrderNumber {
			// increase other progress markers, of which the order number is larger than oldOrderNumber, by one.
			_, err = pg.Exec("UPDATE progress_markers SET order_number = order_number + 1 "+
				"WHERE boundary_partner_id = $1 and order_number >= $2 and order_number <= $3",
				boundaryPartnerId, input.OrderNumber, oldOrderNumber-1)
		} else {
			// decrease other progress markers, of which the order number is smaller than input.OrderNumber, by one.
			JSON(w, http.StatusOK, Response{nil, "-1"})
			_, err = pg.Exec("UPDATE progress_markers SET order_number = order_number - 1 "+
				"WHERE boundary_partner_id = $1 and order_number >= $2 and order_number <= $3",
				boundaryPartnerId, oldOrderNumber+1, input.OrderNumber)
		}
		if err != nil {
			JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
			return
		}
	}

	_, err = pg.Exec("UPDATE progress_markers set title = $1, type = $2, order_number = $3 WHERE progress_marker_id = $4",
		&input.Title, &input.Type, &input.OrderNumber, progressMarkerId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateChallenge(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	challengeId := mux.Vars(r)["challengeId"]
	challengeName := r.FormValue("challenge")

	_, err = pg.Exec("UPDATE challenges SET challenge_name = $1 WHERE challenge_id = $2",
		challengeName, challengeId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}
	JSON(w, http.StatusOK, Response{nil, "success"})
}

func updateStrategy(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	strategyId := mux.Vars(r)["strategyId"]
	strategyName := r.FormValue("strategy")

	_, err = pg.Exec("UPDATE strategies SET strategy_name = $1 WHERE strategy_id = $2",
		strategyName, strategyId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}
	JSON(w, http.StatusOK, Response{nil, "success"})
}

func deleteBoundaryPartner(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	boundaryPartnerId := mux.Vars(r)["partnerId"]
	_, err = pg.Exec("DELETE FROM boundary_partners WHERE boundary_partner_id = $1", boundaryPartnerId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func resetOutcomeStatement(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	boundaryPartnerId := mux.Vars(r)["partnerId"]
	_, err = pg.Exec("UPDATE boundary_partners SET outcome_statement = NULL WHERE boundary_partner_id = $1", boundaryPartnerId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func deleteProgressMarker(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	progressMarkerId := mux.Vars(r)["progressMarkerId"]
	_, err = pg.Exec("DELETE FROM progress_markers WHERE progress_marker_id = $1", progressMarkerId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func deleteChallenge(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	challengeId := mux.Vars(r)["challengeId"]
	_, err = pg.Exec("DELETE FROM challenges WHERE challenge_id = $1", challengeId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func deleteStrategy(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	strategyId := mux.Vars(r)["strategyId"]
	_, err = pg.Exec("DELETE FROM strategies WHERE strategy_id = $1", strategyId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func getExternalResource(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	resources := []ExternalResources{}

	rows, err := pg.Query("SELECT resource_id, resource_url, resource_name from external_resources where project_id = $1", projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	for rows.Next() {
		exr := ExternalResources{}
		err = rows.Scan(&exr.ResourceId, &exr.ResourceUrl, &exr.ResourceName)
		if err != nil {
			rows.Close()
			JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
			return
		}
		resources = append(resources, exr)
	}

	JSON(w, http.StatusOK, Response{resources, "success"})
}

// TODO how to handle an uploaded file that share the same name with an existing file.
func uploadResourceFile(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	r.ParseMultipartForm(10 * int64(1e6)) // grab the multipart form
	if r.MultipartForm == nil {
		JSON(w, http.StatusBadRequest, Response{nil, "request must be multipart/form-data"})
		return
	}

	file, handler, err := r.FormFile("resource_file") // grab the filenames

	if err != nil {
		JSON(w, http.StatusBadRequest, Response{nil, err.Error()})
		return
	}
	defer file.Close()

	url, err := uploadFile(file, handler.Filename, projectId)
	if err != nil {
		JSON(w, http.StatusBadRequest, Response{nil, err.Error()})
		return
	}

	_, err = pg.Exec("INSERT INTO external_resources (resource_id, project_id, resource_url, resource_name, created_by) VALUES ($1, $2, $3, $4, $5)",
		uuid.NewV1().String(), projectId, url, handler.Filename, user.UserId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func uploadFile(file io.ReadSeeker, name string, project_id string) (string, error) {
	if _, err := os.Stat("./temp_file/" + project_id); os.IsNotExist(err) {
		os.Mkdir("./temp_file/"+project_id, 0777)
	}

	f, err := os.OpenFile("./temp_file/"+project_id+"/"+name, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	io.Copy(f, file)
	return "./temp_file/" + project_id + "/" + name, nil
}

func deleteReasourceFile(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, USER).(User)
	if user.IsAdmin == false {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	projectId := mux.Vars(r)["projectId"]
	projectOrganizationId, err := dbGetProjectOrganization(projectId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	// check that the current user and the project is for the same organization
	if projectOrganizationId != user.OrganizationId {
		JSON(w, http.StatusForbidden, Response{nil, "Permission denied"})
		return
	}

	var file_path string

	resourceId := mux.Vars(r)["resourceId"]

	err = pg.QueryRow("SELECT resource_url FROM external_resources WHERE resource_id = $1", resourceId).Scan(&file_path)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	err = os.Remove(file_path)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	_, err = pg.Exec("DELETE FROM external_resources WHERE resource_id = $1", resourceId)
	if err != nil {
		JSON(w, http.StatusInternalServerError, Response{nil, err.Error()})
		return
	}

	JSON(w, http.StatusOK, Response{nil, "success"})
}

func main() {
	userSessions = NewSessionStorage()
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	router.HandleFunc("/", index).Methods(GET)

	// login
	router.HandleFunc("/login", login).Methods(POST)

	// projects
	router.HandleFunc("/projects", authenticate(getProjects)).Methods(GET)
	router.HandleFunc("/projects/add", authenticate(addProject)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/update/project_name", authenticate(updateProjectName)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/update/project_logo", authenticate(updateProjectLogo)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/update/project_description", authenticate(updateProjectDescription)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/update/project_budget", authenticate(updateProjectBudget)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/update/project_timeline", authenticate(updateProjectTimeline)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/update/project_donor", authenticate(updateProjectDonor)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/update/project_mission", authenticate(updateProjectMission)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/update/project_vision", authenticate(updateProjectVision)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/delete/project", authenticate(deleteProject)).Methods(DELETE)
	router.HandleFunc("/projects/{projectId}/reset/project_name", authenticate(resetProjectName)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/reset/project_logo", authenticate(resetProjectLogo)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/reset/project_description", authenticate(resetProjectDescription)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/reset/project_budget", authenticate(resetProjectBudget)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/reset/project_timeline", authenticate(resetProjectTimeline)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/reset/project_donor", authenticate(resetProjectDonor)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/reset/project_mission", authenticate(resetProjectMission)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/reset/project_vision", authenticate(resetProjectVision)).Methods(POST)

	// project boundary partners
	router.HandleFunc("/projects/{projectId}/add_boundary_partner", authenticate(addBoundaryPartner)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/{partnerId}/get", authenticate(getBoundaryPartner)).Methods(GET)
	router.HandleFunc("/projects/{projectId}/{partnerId}/get2", authenticate(getBoundaryPartner1)).Methods(GET)

	// boundary partner's progress markers
	router.HandleFunc("/projects/{projectId}/{partnerId}/add_progress_marker", authenticate(addProgressMarker)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/{partnerId}/{progressMarkerId}/add_challenge", authenticate(addChallenge)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/{partnerId}/{progressMarkerId}/add_strategy", authenticate(addStrategy)).Methods(POST)

	// update boundary partner
	router.HandleFunc("/projects/{projectId}/{partnerId}/update/partner_name", authenticate(updatePartnerName)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/{partnerId}/update/outcome_statement", authenticate(updateOutcomeStatement)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/{partnerId}/{progressMarkerId}/update/progressMarker", authenticate(updateProgressMarker)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/{partnerId}/{challengeId}/update/challenge", authenticate(updateChallenge)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/{partnerId}/{strategyId}/update/strategy", authenticate(updateStrategy)).Methods(POST)

	// delete boundary partner
	router.HandleFunc("/projects/{projectId}/{partnerId}/delete/partner", authenticate(deleteBoundaryPartner)).Methods(DELETE)
	router.HandleFunc("/projects/{projectId}/{partnerId}/reset/outcome_statement", authenticate(resetOutcomeStatement)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/{progressMarkerId}/delete/progress_marker", authenticate(deleteProgressMarker)).Methods(DELETE)
	router.HandleFunc("/projects/{projectId}/{challengeId}/delete/challenge", authenticate(deleteChallenge)).Methods(DELETE)
	router.HandleFunc("/projects/{projectId}/{strategyId}/delete/strategy", authenticate(deleteStrategy)).Methods(DELETE)

	// external resources
	router.HandleFunc("/projects/{projectId}/resource", authenticate(getExternalResource)).Methods(GET)
	router.HandleFunc("/projects/{projectId}/resource_uploadfile", authenticate(uploadResourceFile)).Methods(POST)
	router.HandleFunc("/projects/{projectId}/{resourceId}/delete/resource_file", authenticate(deleteReasourceFile)).Methods(DELETE)

	connectPostgres()
	n := negroni.New()
	n.UseHandler(router)
	n.Run(LISTEN_ADDR)
}
