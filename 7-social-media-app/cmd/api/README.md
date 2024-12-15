Anything related to server and HTTP related.

Organized in 3 layers:

Transport: How we deliver the message, in our case HTTP. Will also contain tests with mocks as it orchestrates tasks.

Service: Where business logic lives. Will also be split into repositories, for example, a user repository that takes care of user logic.

Storage: Abstractions between our database and the layers above so that they don't need to know the underlying tech behind the database
