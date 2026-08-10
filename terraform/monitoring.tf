# CloudWatch Log Group for Application Logs (14-day retention to minimize cost)
resource "aws_cloudwatch_log_group" "app_logs" {
  name              = "/aws/ec2/${var.project_name}"
  retention_in_days = 14

  tags = {
    Name = "${var.project_name}-log-group"
  }
}

# Metric Alarm: High CPU Utilization (> 80% for 5 minutes)
resource "aws_cloudwatch_metric_alarm" "high_cpu" {
  alarm_name          = "${var.project_name}-high-cpu"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "Alarm when EC2 CPU utilization exceeds 80% for 5 minutes"

  dimensions = {
    InstanceId = aws_instance.app_server.id
  }

  tags = {
    Name = "${var.project_name}-high-cpu-alarm"
  }
}

# Metric Alarm: EC2 Status Check Failures
resource "aws_cloudwatch_metric_alarm" "status_check_failed" {
  alarm_name          = "${var.project_name}-status-check-failed"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "StatusCheckFailed"
  namespace           = "AWS/EC2"
  period              = 60
  statistic           = "Maximum"
  threshold           = 1
  alarm_description   = "Alarm when EC2 status check fails"

  dimensions = {
    InstanceId = aws_instance.app_server.id
  }

  tags = {
    Name = "${var.project_name}-status-check-alarm"
  }
}
