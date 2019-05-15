provider "aws" {}

data "aws_iam_policy_document" "write" {
  statement {
    sid    = "ECRGetAuthorizationToken"
    effect = "Allow"

    actions = [
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayer*",
      "ecr:CompleteLayerUpload",
      "ecr:PutImage",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_policy" "write" {
  name        = "testpolicy-deleteme"
  description = "Allow IAM Users to pull from ECR"
  policy      = "${data.aws_iam_policy_document.write.json}"
}
