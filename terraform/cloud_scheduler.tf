resource "google_cloud_scheduler_job" "calendar_notifier" {
  name             = var.name
  schedule         = var.schedule
  time_zone        = "Asia/Tokyo"
  attempt_deadline = "180s"

  http_target {
    http_method = "POST"
    uri         = "${google_cloud_run_service.calendar_notifier.status[0].url}/launch"

    oidc_token {
      service_account_email = var.scheduler_service_account
      audience              = google_cloud_run_service.calendar_notifier.status[0].url
    }
  }

  retry_config {
    max_backoff_duration = "3600s"
    max_doublings        = 16
    max_retry_duration   = "0s"
    min_backoff_duration = "5s"
    retry_count          = 0
  }
}
