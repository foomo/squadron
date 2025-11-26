variable "TAG" {
  type = string
  default = "$TAG"
  description = "Tag to use for build"
}
group "all" {
  targets = ["squadron-storefinder-backend-default"]
}
target "squadron-storefinder-backend-default" {
  args = {
    SQUADRON_NAME      = "storefinder"
    SQUADRON_UNIT_NAME = "backend"
  }
  labels = {
    # test
    # test
    # test
    # test
  }
  tags = ["storefinder/backend:${TAG}"]
}
