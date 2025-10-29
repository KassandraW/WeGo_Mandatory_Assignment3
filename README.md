# How to run the program
## Running the server
Open the project in a terminal. Navigate to the server folder and run the command "go run server.go" to start the server up. 
When you want to terminate the server, you can either type exit in the terminal to properly broadcast the server shutting down, or simply press ctrl + c.
## Running clients
Open another terminal, (or multiple terminals if you want more clients) and do the following: navigate to the client folder and run the command : "go run client.go" to start up a client.

## Client interactions
### Sending messages 
 In the client terminal, simply write your message and hit enter. This sends the message to the Server, where it gets broadcasted to all clients.
### Leaving the chat
Hit "control + c" to exit the chat.

## Log files
When running the server, it generates a clean log file, "Log_info" in ./grpc. We have included a log file that shows the requirements are met  in ./ called System_Logs
