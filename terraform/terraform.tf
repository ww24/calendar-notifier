terraform {
  required_version = "~> 0.14.3"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 3.51.0"
    }
  }

  backend "remote" {
    organization = "ww24"

    workspaces {
      name = "calendar-notifier"
    }
  }
}
