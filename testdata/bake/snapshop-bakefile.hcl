variable "TAG" {
  type = string
  default = "latest"
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
  tags            = ["storefinder/backend:${TAG}"]
  no-cache-filter = ["nocache"]
  secret = [
    {
      id = "GITHUB_TOKEN"
    }
  ]
}
