# Go JWT Authentication API

This is a Go-based RESTful API for user authentication using JSON Web Tokens (JWT). It provides endpoints for user signup, login, user data management, tag management, and search functionality.

## Main

The main package initializes the application by loading environment variables, connecting to the database, and defining API routes.
## Controllers

### Signup

Handles user signup by creating a new user with provided credentials.

### Login

Handles user login by verifying credentials and generating JWT tokens for authentication.

### UserData

Retrieves essential user data including ID, email, name, admin status, and associated tags.

### EditUserData

Allows users to edit their own data including email, password, name, and profile image.

### UsersData

Retrieves data of all users including ID, email, name, and admin status.

### EditUserDataByID

Allows admin users to edit user data by ID.

### CreateTag

Creates a new tag with a given name and image.

### EditTag

Allows editing of tag data including name and image.

### AddTagToUser

Associates a tag with a user.

### GetAllTags

Retrieves all tags from the database with pagination support.

### Search

Performs a search for tags and users based on the provided query.

### IsValidEmail

Validates email format using regular expressions.

### SaveUploadedImage

Saves uploaded images to the server and returns the filename.


## Middleware

This package contains middleware functions used in the authentication and authorization process of the Go JWT Authentication API.

### RequireAuth

Authenticates users via JWT tokens stored in cookies. This middleware extracts the token from the cookie named "Authorization" and verifies its validity. If the token is valid, it sets the user data in the context for further processing. If the token is invalid or expired, it aborts the request with a status of Unauthorized (401).

### RequireAdmin

Checks if the user extracted from the context is an admin. This middleware is typically used after RequireAuth to ensure that only admin users can access certain routes or perform specific actions. If the user is not an admin, it aborts the request with a status of Forbidden (403).

