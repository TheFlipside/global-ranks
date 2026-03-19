# GlobalRanks

This project implements a platform which stores data related to unique user IDs,
mainly targeted as a global game leaderboard.

## Architecture

The core is a PostgreSQL database.

As backend a service written in Go handles the communication between clients and
the database.
Clients which connect to the server are added to the database, together with
a randomly generated username.
For a visual enhancement a user avatar is also generated, using a Identicon
algorithm to generate an image by using the UUID for example.

On client side an implementation is needed which:

- Generates a random UUID and stores it locally on the client.
- Checks upon start if a UUID already exists.
- Communicates with the server on a specific event, mainly when submitting a new
  game score, using the UUID.
- Receives a list of scores from the server to display it together with user
  names and avatars.
- Allows the user to edit the assigned username bound to the UUID.

On the server side multiple endpoints need to be available:

- Send the current data containing a list of scores.
- Accept data sent by clients to process, both for new scores or changes to the
  user ID.

Various checks need to be implemented on the server to prevent tampering or
other malicious actions.

- Sanity checks for submitted scores, e.g. no negative values.
- Rate limit how often a given user ID (and IP) can submit scores
  (e.g. 1 per N seconds, with a small burst window).
  To throttle bots and blind POST spam.

## Installation

For the installation on a server see the instructions in INSTALLATION.md

## API

For information about how to implement the communication in a client system, see
API.md

## Development

See DEVELOPMENT.md for instructions how to set up a testing environment.
