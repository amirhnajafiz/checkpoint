# Checkpoint

Using JWT to authenticate the users of a web-application.

## Routes
- **/api/login**
  - method = POST
  - input: username, password
- **/api/register** 
  - method = POST
  - input: username, password
- **/api/user**
  - method = GET
  - input: username, token

## Monitoring
Using airbrake to monitor our project:
<img src="./airbrake.png" width="700" />

## Features
- JWT
- Auth middleware
- Unit testing
- Airbrake monitoring 

## Tools
- go (1.17)
- jwt-go (3.2.0)
- gobrake (5.2.0)

## How to use?
Clone the project and enter the following command:
```shell
make dev
```
