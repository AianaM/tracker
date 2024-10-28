package main

import "fmt"

// то что понадобится для работы с API Yandex Cloud
type cloud struct {
	// Unique ID of the user Tracker account
	id,
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
	cloud := makeRequestData(
		"https://resource-manager.api.cloud.yandex.net/resource-manager/v1/clouds",
		map[string]string{},
		map[string]string{},
		struct{ organizationId string }{},
	)
	if err := cloud.get(); err != nil {
		fmt.Println(err)
		panic("organizationId is not set")
	}
	return cloud.body.organizationId
}

func getMyselfId(org string) string {
	myself := makeRequestData(
		"https://api.tracker.yandex.net/v2/myself",
		map[string]string{"organizationId": org},
		map[string]string{},
		struct{ id string }{},
	)
	if err := myself.get(); err != nil {
		fmt.Println(err)
		panic("id is not set")
	}
	return myself.body.id
}
