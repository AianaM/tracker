package main

import (
	"fmt"
	"log"
)

var c cloud

// то что понадобится для работы с API Yandex Cloud
type cloud struct {
	// Unique ID of the user Tracker account
	id int
	// organizationId пользователя в Yandex Cloud
	orgId string
}

func makeClouds() cloud {
	org := getCloudsOrg()
	id := getMyselfId(org)
	return cloud{
		id:    id,
		orgId: org,
	}
}

func getCloudsOrg() string {
	cloud := newRequestData(
		"https://resource-manager.api.cloud.yandex.net/resource-manager/v1/clouds",
		map[string]string{},
		nil,
		struct {
			Clouds []struct {
				ID             string `json:"id"`
				CreatedAt      string `json:"createdAt"`
				Name           string `json:"name"`
				OrganizationID string `json:"organizationId"`
			} `json:"clouds"`
		}{},
	)
	if err := cloud.get(); err != nil {
		fmt.Println(err)
		panic("organizationId is not set")
	}
	log.Println(cloud.body)
	return cloud.body.Clouds[0].OrganizationID
}

func getMyselfId(org string) int {
	myself := newRequestData(
		"https://api.tracker.yandex.net/v2/myself",
		map[string]string{"X-Cloud-Org-ID": org},
		nil,
		struct {
			Uid int `json:"uid"`
		}{},
	)
	if err := myself.get(); err != nil {
		fmt.Println(err)
		panic("id is not set")
	}
	return myself.body.Uid
}
