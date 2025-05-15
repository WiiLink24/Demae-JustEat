# Just Eat API

This folder contains an implementation of the Just Eat API 
for use with the Demae Channel.

## Important Files
- #### client.go: Initializer for the JEClient. Implements the Client interface for when/if Skip and Grubhub are implemented
- #### server.go: Web server for finalizing the Just Eat order. Required as we need the user to pay for the food using PayPal.
- #### braintree.go: PayPal's E-Commerce infrastructure that Just Eat uses.

## Credits:
- giustino: Reverse engineering Just Eat's APIs
- Noah: Implementing Just Eat's APIs into a Demae Channel recognizable format.
