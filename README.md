# funny-citrus
Basic Golang backend to navigate a folder

## Uploading Rules
1. You can upload only if specifying a 'project'. if a resource with the same name inside the same project already exists the upload fails.
2. you can create projects by posting at /{project}. If said project already exist the upload fails.
3. you can GET the whole project through the /{project} endpoint. The page will be constructed based on the "template.md" in the root of the folder
4. you can GET specific resources with /{project}/{resource}.
5. you can DELETE a resource or project with the appropriate URL (/{project} or /{project}/{resource}).
6. you can access the main page at "/", which will prompt you to "/login" then show you the projects list.

special rules
- /projects resturns a JSON with an array of projects, no porject can be named "projects"
- /{project}/download downloads the whole project as a zip, no resource can be named "download"

### Sources
This is a [wonderful golang toutorial](https://tutorialedge.net/golang/creating-simple-web-server-with-golang/) to do something like this project