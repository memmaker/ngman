# ngman

A simple CLI tool for managing nginx sites.

Will create a JSON file to reflect the current site setup.

Only supports these basic operations:

1. Create a new site with a specified domain name
2. Add a static resource location
3. Add a reverse proxy location
4. Call a custom script for SSL certificate generation

Well, actually, you can also list all sites and delete a site.
That's all folks.

