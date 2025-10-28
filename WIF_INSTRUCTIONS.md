# How to Use Workload Identity Federation for Local Development

This guide provides instructions on how to use Workload Identity Federation (WIF) to authenticate with Google Cloud services from your local machine. This method is more secure than using service account keys and is the recommended approach for local development.

## Prerequisites

*   You have a Google Cloud project.
*   You have the `gcloud` command-line tool installed on your local machine.
*   Your project has a Workload Identity Pool and Provider configured for GitHub Actions.

## Steps

1.  **Log in with your user account.**
    *   First, ensure you are logged into the `gcloud` CLI with your own user account, which has permissions to impersonate the target service account.
    *   Run the following command. This will open a browser window for you to log in.
    ```bash
    gcloud auth login
    ```

2.  **Generate the WIF Credential Configuration File.**
    *   Run the following `gcloud` command to generate the correct credential configuration file. This command tells `gcloud` to create a configuration that allows your local user credentials to impersonate a service account through your Workload Identity Provider.

    *   **Replace the placeholders before running:**
        *   `SERVICE_ACCOUNT_EMAIL`: The email of the service account used by the GitHub workflow.
        *   `FULL_WIF_PROVIDER_ID`: The full identifier of the Workload Identity Provider (e.g., `projects/12345/locations/global/workloadIdentityPools/my-pool/providers/my-repo-provider`).
        *   `/path/to/your/wif-config.json`: The location where you want to save the generated file.

    ```bash
    gcloud iam workload-identity-pools create-cred-config \
        "FULL_WIF_PROVIDER_ID" \
        --service-account="SERVICE_ACCOUNT_EMAIL" \
        --output-file="/path/to/your/wif-config.json" \
        --credential-source-type="user-creds"
    ```

3.  **Find your WIF Provider and Service Account Email.**
    *   If you don't have these values, you can find them in the Google Cloud Console:
    *   **Service Account Email:**
        1.  Navigate to **IAM & Admin** > **Service Accounts**.
        2.  Find the service account used by your workflow and copy its email.
    *   **Workload Identity Provider ID:**
        1.  Navigate to **IAM & Admin** > **Workload Identity Federation**.
        2.  Select the correct pool.
        3.  The full provider ID is listed in the providers table.

4.  **Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable.**
    *   Open a terminal and set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to the **absolute path** of the `wif-config.json` file you just created.
    ```bash
    export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/wif-config.json"
    ```
    *   To make this change permanent, add the `export` command to your shell's startup file (e.g., `~/.bashrc`, `~/.zshrc`).

5.  **Verify your authentication.**
    *   Now, when you run `gcloud` commands, they will automatically use the configuration file to impersonate the service account.
    *   Test this by running a command that requires permission, for example:
    ```bash
    gcloud artifacts docker images list europe-west1-docker.pkg.dev/sbcwaste/eu-docker-repo --project=sbcwaste
    ```
    *   This command should now execute successfully as the service account.
