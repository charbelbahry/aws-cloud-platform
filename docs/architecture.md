# Architecture Design & Technical Decisions

This document outlines the architectural trade-offs, security model, networking design, and infrastructure decisions for the **aws-cloud-platform** project.

---

## 1. Architectural Decisions

### Go Standard Library (`net/http`) vs Third-Party Frameworks
* **Decision:** We used the Go 1.22+ standard library `net/http` router instead of frameworks like Gin or Echo.
* **Rationale:** Go 1.22 introduced method matching (`POST /services`) and path parameters (`/services/{id}`) directly into `http.ServeMux`. Eliminating third-party web frameworks reduces binary size, minimizes external security vulnerabilities, and improves execution speed.

### OpenTofu / Terraform vs CloudFormation / AWS CDK
* **Decision:** Provisioning infrastructure as code using OpenTofu 1.12+.
* **Rationale:** OpenTofu provides open-source, vendor-neutral declarative HCL configuration with state locking and plan verification.

### AWS Free Tier Architecture & NAT Gateway Omission
* **Decision:** Omitted NAT Gateways (~$32/month cost) from the VPC design.
* **Rationale:** Placing EC2 in a public subnet with Internet Gateway (IGW) routing while keeping RDS isolated in private subnets guarantees **100% Free Tier ($0/month)** compliance while enforcing database isolation.

---

## 2. Security Model

1. **Database Isolation:** PostgreSQL RDS instance is in private subnets without public IPs assigned.
2. **Stateful SG-to-SG Firewalls:** Database Security Group (`db-sg`) permits TCP port 5432 ingress **only from `app-sg`**.
3. **IAM Instance Profiles:** EC2 authenticates to Amazon ECR and CloudWatch using short-lived STS tokens via IAM instance profile, eliminating static AWS keys on disk.
4. **Least-Privilege Container:** Multi-stage Docker build runs as non-root UID `1000:1000`.

---

## 3. Database Migration Strategy
SQL migrations are embedded in the Go binary using `//go:embed *.sql` and executed inside atomic database transactions on server boot, ensuring zero external tool dependencies.
