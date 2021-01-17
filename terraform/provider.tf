provider "google" {
  credentials = var.google_credentials
  project     = var.project
  region      = var.location
}
