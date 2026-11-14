data "aws_iam_policy_document" "assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "artifact_read" {
  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.artifact.arn}/${aws_s3_object.artifact.key}"]
  }
}

resource "aws_iam_role" "site" {
  name               = "${var.project}-role"
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

resource "aws_iam_role_policy" "artifact_read" {
  name   = "${var.project}-artifact-read"
  role   = aws_iam_role.site.id
  policy = data.aws_iam_policy_document.artifact_read.json
}

resource "aws_iam_role_policy_attachment" "ssm" {
  role       = aws_iam_role.site.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "site" {
  name = "${var.project}-profile"
  role = aws_iam_role.site.name
}
