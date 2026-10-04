# funny-citrus
Basic Golang BACKEND to navigate a project-organized directory (just visualization)

## Uploading Rules (none)
1. You don't upload


## Key-Rules
0. only GET requests allowed.
1. only GET requests have standard pathing (e.g. /p/my-project/awesome-footpic.png, /index.html)
2. all projects related endpoints are under "/p" route.

## Endpoints
### Special Endpoints
- `GET /p`: resturns a JSON with a list of all projects. All project related actions will be under this route
- `GET /p/download?project=<project>`: downloads a project (zip'ed folder)
### Standard Endpoints
- `GET /`: index.html, all static assets under this route
- `GET /p/{project}`: returns a JSON found in the project directory:
```json
    {
        "description": "Brief description",         // Description of the project
        "resources": [                              // Array of resources
            {
                "filename": "secret code",          // name of the file
                "time_of_upload": "tomorrow",       // when was uploaded
                "size": "so fat (bytes)"               // size of the file (number)
            }
        ],
        "res_count": "it's over 9000",              // how many resources in the project (number)
        "project_size": "really fat (bytes)"        // size of the whole project (number)
    }
```
- `GET /p/{project}/{resource}` get the specific resource

## Visualization
VueJS frontend for Homepage and each Project folder [happy-citrus](https://github.com/Bretzz/happy-citrus)

### Sources
This is a [wonderful golang toutorial](https://tutorialedge.net/golang/creating-simple-web-server-with-golang/) to do something like this project