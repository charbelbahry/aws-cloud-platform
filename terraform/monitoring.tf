# CloudWatch Log Group for Application Logs (14-day retention to minimize cost)
resource "aws_cloudwatch_log_group" "app_logs" {
  name              = "/aws/ec2/${var.project_name}"
  retention_in_days = 14

  tags = {
    Name = "${var.project_name}-log-group"
  }
}

# Metric Filter to parse ERROR level logs from Go slog JSON output
resource "aws_cloudwatch_log_metric_filter" "app_errors" {
  name           = "${var.project_name}-app-errors"
  pattern        = "{ $.level = \"ERROR\" }"
  log_group_name = aws_cloudwatch_log_group.app_logs.name

  metric_transformation {
    name      = "AppErrorCount"
    namespace = "AWSCloudPlatform/Application"
    value     = "1"
  }
}

# CloudWatch Alarm for Go Application Errors
resource "aws_cloudwatch_metric_alarm" "app_errors" {
  alarm_name          = "${var.project_name}-app-error-alarm"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = aws_cloudwatch_log_metric_filter.app_errors.metric_transformation[0].name
  namespace           = aws_cloudwatch_log_metric_filter.app_errors.metric_transformation[0].namespace
  period              = 300
  statistic           = "Sum"
  threshold           = 1
  alarm_description   = "Alarm when Go application logs 1 or more ERROR level logs"
  treat_missing_data  = "notBreaching"

  tags = {
    Name = "${var.project_name}-app-error-alarm"
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
