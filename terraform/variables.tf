variable "location" {
  type    = string
  default = "asia-northeast1"
}

variable "project" {
  type = string
}

// credentials json value
variable "google_credentials" {
  type = string
}

variable "name" {
  type    = string
  default = "calendar-notifier"
}

variable "gar_repository" {
  type    = string
  default = "ww24"
}

variable "image_name" {
  type    = string
  default = "calendar-notifier"
}

variable "image_tag" {
  type    = string
  default = "latest"
}

// cloud run service account
variable "service_account" {
  type = string
}

// cloud scheduler service account
variable "scheduler_service_account" {
  type = string
}

variable "schedule" {
  type    = string
  default = "*/10 * * * *"
}

// calendar-notifier config (yaml)
variable "config" {
  type = string
}
