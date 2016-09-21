.. highlight:: http

.. contents:: Table of Contents
    :backlinks: none
    :depth: 4

Lucid Web App API Documentation
===============================

This document follows RFC2119_ conventions for keywords.

API base URL: https://someurl.com/ need to set up server

The server operates over HTTPS *only*. Non-secure HTTP is not supported and
*must* not be used.


API Server Protocol
-------------------

- All API endpoints return JSON-serialized data which *must* contain the
  following keys:

  - ``data: any`` - response data, may be any value including ``null``
  - ``message: string`` - human-readable message for debugging purposes

- All JSON keys follow the snake_case format

The format of the ``user`` object is as follows:

.. code-block:: json

    {
        "user_id": "8c1cbd35-8ad9-482b-b593-81a5bd1b5146",
        "full_name": "First Last"
    }

- ``user_id`` is a RFC4122_ UUID string. The special value ``me`` is an alias
  for the user UUID of the currently logged in user.

- Any timestamps returned by the server are standard Unix timestamps.


Authentication
--------------

All authenticated endpoints require the client to set the ``X-Api-Key`` HTTP
header to the key received during the authentication process.

HTTP Endpoints
--------------

Login
^^^^^

.. http:post:: /login

    This endpoint logs in a user with the supplied ``email`` and ``password``.

    The ``data`` field contains the API key.

    :status 401: if authentication fails

    **Example request**::

        POST /v1/login HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        email=...
        password=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": "tuYeHXqeK8yMe+1GaSsOsdVOfSdevPoddfP8vh46ath=",
            "message": "logged in"
        }

Projects
^^^^^^^^

.. http:get:: /projects

    This gets the list of projects belonging to the user's organization.

    :reqheader X-Api-Key: required API key

    **Example request**::

        GET /projects HTTP/1.1

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": [
                {
                    "project_id": "9ac2ee6c-f2b0-4537-bda7-6c5057109f87",
                    "project_name": "My Project Name",
                    "logo_url": "",
                    "description": "My project description",
                    "budget": 100000,
                    "donor": "",
                    "vision":"",
                    "mission":"",
                    "timeline_from":"2012-11-01 10:08:41 UTC",
                    "timeline_to":"2012-12-21 03:08:41 UTC",
                    "boundary_partner_ids": null,
                    "boundary_partner_names": null,
                    "resource_ids": null,
                    "resource_urls": null
                }
            ],
            "message": "success"
        }

.. http:post:: /projects/add

    This endpoint adds a new project.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        GET /projects HTTP/1.1
        Content-Type: application/json

        {
            "project_name": "New Project Name",
            "description": "New project description",
            "budget": 50000
        }

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/update/project_name

    This endpoint updates the project's name.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/update/project_name HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        project_name=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/update/project_logo

    This endpoint updates the project's logo.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/update/project_logo HTTP/1.1
        Content-Type: multipart/form-data

        project_logo = ...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/update/project_description

    This endpoint updates the project's description

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/update/project_description HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        project_description=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/update/project_budget

    This endpoint updates the project's budget

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/update/project_budget HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        project_budget=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/update/project_timeline

    This endpoint updates the project's timeline. Both `timeline_from` and `timeline_to` is in `RFC3339` format.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/update/project_timeline HTTP/1.1
        Content-Type: application/json

        "timeline_from":"2012-11-01T22:08:41+00:00",
        "timeline_to":"2012-12-21T15:08:41+00:00"'

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/update/project_donor

    This endpoint updates the project's donor

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/update/project_donor HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        project_donor=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/update/project_mission

    This endpoint updates the project's mission.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/update/project_mission HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        project_mission=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/update/project_vision

    This endpoint updates the project's vision.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/update/project_vision HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        project_vision=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:delete:: /projects/{projectId}/delete/project

    This endpoint deletes the project and related boundary partners, external resources.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        DELETE /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/delete/project HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/reset/project_name

    This endpoint resets the project's name.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/reset/project_name HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/reset/project_logo

    This endpoint resets the project's logo

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin.

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/reset/project_logo HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/reset/project_description

    This endpoint resets the project's description.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/reset/project_description HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/reset/project_budget

    This endpoint resets the project's budget.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/reset/project_budget HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/reset/project_timeline

    This endpoint resets the project's timeline.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/reset/project_timeline HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/reset/project_donor

    This endpoint resets the project's donor.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/reset/project_donor HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/reset/project_mission

    This endpoint resets the project's mission.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/reset/project_mission HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/reset/project_vision

    This endpoint resets the project's vision.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/reset/project_vision HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }


Projects Boundary Partners
^^^^^^^^^^^^^^^^^^^^^^^^^^

.. http:post:: /projects/{projectId}/add_boundary_partner

    This endpoint adds a new boundary partner for the project.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/add_boundary_partner HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        partner_name=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:get:: /projects/{projectId}/{partnerId}/get

    This endpoint gets the boundary partner's info including the progress markers, challenges and strategies.

    :reqheader X-Api-Key: required API key

    **Example request**::

        GET /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/get HTTP/1.1

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": {
                "boundary_partner_id": "24e27933-4b3c-40ee-bfee-18db5fb96419",
                "project_id": "9ac2ee6c-f2b0-4537-bda7-6c5057109f87",
                "partner_name": "Joe Kennings",
                "outcome_statement": "This is my outcome statement",
                "progress_markers": [
                    {
                        "progress_marker_id": "e1a738ca-892b-4927-b3c3-103879fe807c",
                        "boundary_partner_id": "24e27933-4b3c-40ee-bfee-18db5fb96419",
                        "title": "1st Progress Marker",
                        "type": 0,
                        "order_number": 1,
                        "challenges": [
                            {
                                "challenge_id": "781387a4-0782-4b03-8d8c-99e194971152",
                                "progress_marker_id": "e1a738ca-892b-4927-b3c3-103879fe807c",
                                "challenge_name": "Challenge 1"
                            },
                            {
                                "challenge_id": "c3f8b82a-341f-4e0c-9dd5-10fe672cb83d",
                                "progress_marker_id": "e1a738ca-892b-4927-b3c3-103879fe807c",
                                "challenge_name": "Challenge 2"
                            }
                        ],
                        "strategies": [
                            {
                                "strategy_id": "715d4f03-4b81-4bad-a347-49aeb75102f6",
                                "progress_marker_id": "e1a738ca-892b-4927-b3c3-103879fe807c",
                                "strategy_name": "Strategy 1"
                            },
                            {
                                "strategy_id": "715d4f03-4b81-4bad-a347-49aeb75102f7",
                                "progress_marker_id": "e1a738ca-892b-4927-b3c3-103879fe807c",
                                "strategy_name": "Strategy 2"
                            }
                        ]
                    },
                    {
                        "progress_marker_id": "ac04fcbb-2ba6-4d03-8337-874dc703cd39",
                        "boundary_partner_id": "24e27933-4b3c-40ee-bfee-18db5fb96419",
                        "title": "2nd Progress Marker",
                        "type": 0,
                        "order_number": 2,
                        "challenges": null,
                        "strategies": [
                            {
                                "strategy_id": "646bbbb7-65b6-493c-9e6f-5d8c0fdf6517",
                                "progress_marker_id": "ac04fcbb-2ba6-4d03-8337-874dc703cd39",
                                "strategy_name": "Strategy 1"
                            }
                        ]
                    }
                ]
            },
            "message":"success"
        }

.. http:post:: /projects/{projectId}/{partnerId}/add_progress_marker

    This endpoint adds a new progress marker to the boundary partner. ``type`` is an integer value: 0 - "Expect to see", 1 - "Like to See", and 2 - "Love to See".

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/add_progress_marker HTTP/1.1
        Content-Type: application/json

        {
            "title": "Progress Marker Title",
            "type": 0
        }

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/{partnerId}/{progressMarkerId}/add_challenge

    This endpoint adds a new challenge to the progress marker.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/D0E6A977-8BBF-43C7-B7F1-1D521D4BB71E/add_challenge HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        challenge=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/{partnerId}/{progressMarkerId}/add_strategy

    This endpoint adds a new strategy to the progress marker.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/D0E6A977-8BBF-43C7-B7F1-1D521D4BB71E/add_strategy HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        strategy=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/{partnerId}/update/partner_name

    This endpoint updates the boundary partner's name.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/update/partner_name HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        partner_name=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/{partnerId}/update/outcome_statement

    This endpoint updates the boundary partner outcome statement

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/update/outcome_statement HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        outcome_statement=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/{partnerId}/{progressMarkerId}/update/progressMarker

    This endpoint updates the progress marker's title and type. ``type`` is an integer value: 0 - "Expect to see", 1 - "Like to See", and 2 - "Love to See".

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/25a325e4-d4e5-46c9-bc47-60e00df7bda6/update/progressMarker HTTP/1.1
        Content-Type: application/json

        {
            "title": "Progress Marker Title",
            "type": 0,
            "order_number": 2
        }

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/{partnerId}/{challengeId}/update/challenge

    This endpoint updates the challenge title in a progress marker.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/876763fd-1b42-4f3a-aa2b-d4162aaecf46/update/challenge HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        challenge=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/{partnerId}/{strategyId}/update/strategy

    This endpoint updates the strategy title in a progress marker.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/e259d9e8-a70a-45ec-94a6-f2fb5d550f51/update/strategy HTTP/1.1
        Content-Type: application/x-www-form-urlencoded

        strategy=...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:delete:: /projects/{projectId}/{partnerId}/delete/partner

    This endpoint deletes the boundary partner and related progress marker, challenges and strategies.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        DELETE /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/delete/partner HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:post:: /projects/{projectId}/{partnerId}/reset/outcome_statement

    This endpoint resets the boundary partner overview outcome statement.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/reset/outcome_statement HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:delete:: /projects/{projectId}/{progressMarkerId}/delete/progress_marker

    This endpoint deletes the progress marker and related challenges and strategies.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        DELETE /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/delete/progress_marker HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }


.. http:delete:: /projects/{projectId}/{challengeId}/delete/challenge

    This endpoint deletes the challenge.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        DELETE /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/delete/challenge HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:delete:: /projects/{projectId}/{strategyId}/delete/strategy

    This endpoint deletes the strategy.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        DELETE /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/24e27933-4b3c-40ee-bfee-18db5fb96419/delete/strategy HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:GET:: /projects/{projectId}/resource

    This endpoint gets external resources of the project.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        GET /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/resources HTTP/1.1


    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": [{
                "resource_id":"e0b10a0c-76c0-11e6-b432-80e6501cdff2",
                "project_id":"",
                "resource_url":"",
                "resource_name":"diagram_img.png"
                },
                {
                "resource_id":"f23e9ab7-76c0-11e6-b432-80e6501cdff2",
                "project_id":"","resource_url":"",
                "resource_name":"project_overview.png"}],
            "message": "success"
        }

.. http:POST:: /projects/{projectId}/resource_uploadfile

    This endpoint upload an external resource file to the project.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        POST /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/resources_uploadfile HTTP/1.1
        Content-Type: multipart/form-data

        resource_file = ...

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }

.. http:DELETE:: /projects/{projectId}/{resourceId}/delete/resource_file

    This endpoint deletes an external resource file from the project.

    :reqheader X-Api-Key: required API key
    :status 403: if current user is not an admin

    **Example request**::

        DELETE /projects/9ac2ee6c-f2b0-4537-bda7-6c5057109f87/8d93ffe5-2773-433d-b37c-9a8b374f6abd/delete/resource_file HTTP/1.1

    **Example response**::

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "data": null,
            "message": "success"
        }


.. _RFC2119: https://www.ietf.org/rfc/rfc2119.txt
.. _RFC4122: https://www.ietf.org/rfc/rfc4122.txt
