package main

import "github.com/lib/pq"

type User struct {
	UserId         string `json:"user_id"`
	OrganizationId string `json:"organization_id"`
	FullName       string `json:"full_name"`
	IsAdmin        bool   `json:"is_admin"`
}

type Project struct {
	ProjectId            string         `json:"project_id"`
	ProjectName          string         `json:"project_name"`
	LogoUrl              string         `json:"logo_url"`
	Description          string         `json:"description"`
	Budget               float64        `json:"budget"`
	Donor                string         `json:"donor"`
	Vision               string         `json:"vision"`
	Mission              string         `json:"mission"`
	TimelineFrom         string         `json:"timeline_from"`
	TimelineTo           string         `json:"timeline_to"`
	BoundaryPartnerIds   pq.StringArray `json:"boundary_partner_ids"`
	BoundaryPartnerNames pq.StringArray `json:"boundary_partner_names"`
	ResourceIds          pq.StringArray `json:"resource_ids"`
	ResourceUrls         pq.StringArray `json:"resource_urls"`
}

type ExternalResources struct {
	ResourceId   string `json:"resource_id"`
	ProjectId    string `json:"project_id"`
	ResourceUrl  string `json:"resource_url"`
	ResourceName string `json:"resource_name"`
}

type BoundaryPartner struct {
	BoundaryPartnerId string            `json:"boundary_partner_id"`
	ProjectId         string            `json:"project_id"`
	PartnerName       string            `json:"partner_name"`
	OutcomeStatement  string            `json:"outcome_statement"`
	ProgressMarkers   []*ProgressMarker `json:"progress_markers"`
}

type ProgressMarker struct {
	ProgressMarkerId  string       `json:"progress_marker_id"`
	BoundaryPartnerId string       `json:"boundary_partner_id"`
	Title             string       `json:"title"`
	Type              int          `json:"type"`
	OrderNumber       int          `json:"order_number"`
	Challenges        []*Challenge `json:"challenges"`
	Strategies        []*Strategy  `json:"strategies"`
}

type Challenge struct {
	ChallengeId      string `json:"challenge_id"`
	ProgressMarkerId string `json:"progress_marker_id"`
	ChallengeName    string `json:"challenge_name"`
}

type Strategy struct {
	StrategyId       string `json:"strategy_id"`
	ProgressMarkerId string `json:"progress_marker_id"`
	StrategyName     string `json:"strategy_name"`
}
