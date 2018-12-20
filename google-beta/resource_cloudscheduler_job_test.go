package google

import "fmt"

func testCloudSchedulerJob_pubSub(name string) string {
	return fmt.Sprintf(`

resource "google_pubsub_topic" "topic" {
	name = "build-triggers"
}

resource "google_cloudscheduler_job" "job" {
	name     = "%s"
	schedule = "*/2 * * * *"

	pubsub_target = {
		topic_name = "${google_pubsub_topic.topic.name}"
	}
}

	`, name)
}

func testCloudSchedulerJob_appEngine(name string) string {
	return fmt.Sprintf(`

resource "google_cloudscheduler_job" "job" {
	name     = "%s"
	schedule = "*/4 * * * *"

	// TODO defaults to the default service, investigation required
	app_engine_http_target = {
		relative_uri = "/ping"
	}
}

	`, name)
}

func testCloudSchedulerJob_http(name string) string {
	return fmt.Sprintf(`

resource "google_cloudscheduler_job" "job" {
	name     = "%s"
	schedule = "*/8 * * * *"

	http_target = {
		uri = "https://example.com/ping"
	}
}

	`, name)
}
