# Installation

Welcome to the Nixopus installation guide. This section will help you set up Nixopus on your VPS quickly.

To install Nixopus on your VPS, ensure you have sudo access and run the following command:

```
sudo bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)"
```

## Optional Parameters

You can customize your installation by providing the following optional parameters:

- `--api-domain`: Specify the domain where the Nixopus API will be accessible (e.g., `nixopusapi.example.tld`)
- `--app-domain`: Specify the domain where the Nixopus app will be accessible (e.g., `nixopus.example.tld`)
- `--email` or `-e`: Set the email for the admin account
- `--password` or `-p`: Set the password for the admin account

Example with optional parameters:
```
sudo bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)" -- \
  --api-domain nixopusapi.example.tld \
  --app-domain nixopus.example.tld \
  --email admin@example.tld \
  --password Adminpassword@123 \
  --env production
```

## Accessing Nixopus

After successful installation, you can access the Nixopus dashboard by visiting the URL you specified in the `--app-domain` parameter (e.g., `https://nixopus.example.tld`). Use the email and password you provided during installation to log in.

> **Note**: The installation script has not been tested in all distributions and different operating systems. If you encounter any issues during installation, please create an issue on our [GitHub repository](https://github.com/raghavyuva/nixopus/issues) with details about your environment and the error message you received.

