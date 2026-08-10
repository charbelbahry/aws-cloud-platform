terraform {
  backend "s3" {
    bucket         = "aws-cloud-platform-tfstate-225989356670"
    key            = "dev/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "aws-cloud-platform-tflock"
    encrypt        = true
  }
}
