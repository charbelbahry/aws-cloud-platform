# Find latest Amazon Linux 2023 AMI
data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Key Pair for SSH access
resource "aws_key_pair" "deployer" {
  key_name   = "${var.project_name}-key"
  public_key = file("~/.ssh/id_rsa.pub")
}

# EC2 Instance (Free Tier Eligible)
resource "aws_instance" "app_server" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.app_sg.id]
  iam_instance_profile   = aws_iam_instance_profile.ec2_profile.name
  key_name               = aws_key_pair.deployer.key_name

  user_data = templatefile("${path.module}/user_data.tftpl", {
    ecr_url      = aws_ecr_repository.app.repository_url
    rds_endpoint = aws_db_instance.postgres.endpoint
    db_user      = var.db_username
    db_password  = random_password.db_password.result
    db_name      = var.db_name
  })

  user_data_replace_on_change = true

  tags = {
    Name = "${var.project_name}-server"
  }
}
