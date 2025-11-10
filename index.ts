import * as pulumi from "@pulumi/pulumi";
import * as gcp from "@pulumi/gcp";

// Get configuration
const config = new pulumi.Config();
const projectId = config.get("projectId") || gcp.config.project || "your-gcp-project-id";
const region = config.get("region") || "us-central1";
const deepseekApiUrl = config.require("deepseekApiUrl");
const deepseekApiKey = config.requireSecret("deepseekApiKey");

// Create a Cloud Run service
const service = new gcp.cloudrun.Service("cloud-inference-service", {
    location: region,
    project: projectId,
    template: {
        spec: {
            containers: [{
                image: pulumi.interpolate`gcr.io/${projectId}/cloud-inference:latest`,
                ports: [{
                    containerPort: 8080,
                }],
                envs: [
                    {
                        name: "DEEPSEEK_API_URL",
                        value: deepseekApiUrl,
                    },
                    {
                        name: "DEEPSEEK_API_KEY",
                        value: deepseekApiKey,
                    },
                ],
                resources: {
                    limits: {
                        memory: "512Mi",
                        cpu: "1000m",
                    },
                },
            }],
            containerConcurrency: 80,
            timeoutSeconds: 300,
        },
        metadata: {
            annotations: {
                "autoscaling.knative.dev/minScale": "0",
                "autoscaling.knative.dev/maxScale": "10",
            },
        },
    },
    traffics: [{
        percent: 100,
        latestRevision: true,
    }],
});

// Allow unauthenticated access (or configure IAM as needed)
const iamBinding = new gcp.cloudrun.IamMember("cloud-inference-service-iam", {
    service: service.name,
    location: service.location,
    role: "roles/run.invoker",
    member: "allUsers",
});

// Export the service URL
export const serviceUrl = service.statuses[0].url;
export const serviceName = service.name;

