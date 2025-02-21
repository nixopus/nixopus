# opus

Users will arrive at the landing page, see the "Try Now" action button, and be redirected to the documentation page where they can find a one-line installer for the application. If customization is needed, users can modify the Docker Compose file.

Once the application is installed on the server, it will display the endpoint where the application is accessible.

**Features to Consider:**

1. Support for multiple user logins
2. Role-based access control
3. Connection to multiple servers
4. Separate build server for building applications
5. Separate run server for running applications
6. Option to push images to Docker or custom registries
7. Integration with GitHub and GitLab (support for other alternatives if possible)
8. Real-time display of logs for deployed applications
9. Real-time display of build logs
10. Real-time metrics for deployed Docker applications
11. Display of available images on the server with one-click deployment
12. One-click deployment of images pulled from Docker registry
13. Status updates for applications (e.g., running, stopped, deployed, not deployed, not running, pending)
14. Automatic cleanup of builds and applications when not used for a certain period
15. Ability to connect multiple domain names to an application
16. Automatic proxy configurations and setup, including SSL certificates