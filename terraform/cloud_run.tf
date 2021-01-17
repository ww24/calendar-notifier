data "google_cloud_run_service" "calendar_notifier" {
  name     = var.name
  location = var.location
}

locals {
  current_image = data.google_cloud_run_service.calendar_notifier.template != null ? data.google_cloud_run_service.calendar_notifier.template[0].spec[0].containers[0].image : null
  new_image     = "${var.location}-docker.pkg.dev/${var.project}/${var.gar_repository}/${var.image_name}:${var.image_tag}"
  image         = (local.current_image != null && var.image_tag == "latest") ? local.current_image : local.new_image
}

resource "google_cloud_run_service" "calendar_notifier" {
  name     = var.name
  location = var.location

  template {
    spec {
      service_account_name = var.service_account

      containers {
        image = local.image

        resources {
          limits = {
            cpu    = "1000m"
            memory = "128Mi"
          }
        }

        env {
          name  = "CONFIG"
          value = var.config
        }
      }
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" = "1"
      }

      labels = {
        service = var.name
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}
