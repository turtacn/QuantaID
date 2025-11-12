# Operations Runbook

This runbook provides instructions for common operational tasks.

## Daily Checks

- **Check Logs:** Review the logs for the QuantaID pods for any errors or warnings.
- **Check Monitoring:** Review the Grafana dashboard for any anomalies in the key metrics.

## Scaling

- **Manual Scaling:** To manually scale the number of pods, use the following command:

```bash
kubectl scale deployment <deployment-name> --replicas=<number-of-replicas>
```

- **Automatic Scaling:** The Horizontal Pod Autoscaler is configured to automatically scale the number of pods based on CPU utilization.

## Certificate Renewal

This deployment uses Let's Encrypt to automatically renew TLS certificates. The certificates are valid for 90 days.

## Security Incident Response

- **Suspicious Login:** If you identify a suspicious login, you can lock the user's account by setting the `locked` field in the `users` table to `true`.
- **Data Breach:** In the event of a data breach, follow the company's incident response plan.
