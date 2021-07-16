Python SDK, a python driver for Mainflux HTTP API.

Does both system administration (provisioning) and messaging.

Installation
Import "github.com/mainflux/mainflux/sdk/python" in your Python package.

import "github.com/mainflux/mainflux/pkg/sdk/python"```

Then call SDK Pythong functions to interact with the system.

## API Reference

```python
FUNCTIONS

class SDK:
    def __init__(
        self,
        users_url=default_url,
        things_url=default_url,
        messages_url=default_url,
        groups_url=default_url
    ):
        self.users = users.Users(users_url)
        self.things = things.Things(things_url)
        self.messages = messages.Messages(messages_url)
        self.channels = channels.Channels(things_url)
        self.groups = groups.Groups(groups_url)
        self.version_url = things_url

    def version(self):
        Version - server health check
        
class Users:
    def __init__(self, url):
        self.url = url

    def create(self, user):
        CreateUser - create user

    def login(self, user):
        CreateToken - create user token
    
    def get(self, user_id, token):
        User - gets user

    def get_all(self, token):
        Users - gets all users
    
    def update(self, user, token):
        UpdateUser - update user

    def update_password(self, old_password, password, token):
        UpdatePassword - update user password

class Things:
    def __init__(self, url):
        self.url = url

    def create(self, thing, token):
        CreateThing - creates new thing and generates thing UUID

    def construct_query(self, params):
        ConstructQuery - allows you to set query params: offset, limit, order, direction

    def create_bulk(self, things, token):
        CreateThings - creates bulk of things and generates things UUID

    def get(self, thing_id, token):
        Thing - gets thing
    
    def get_all(self, token, query_params=None):
        Things - gets all things

    def get_by_channel(self, channel_id, params, token):
        GetByChannel - Gets all things to which a specific thing is connected to

    def update(self, thing_id, token, thing):
        UpdateThing - updates thing by ID

    def delete(self, thing_id, token):
        DeleteThing - removes thing

    def connect(self, channel_ids, thing_ids, token):
        Connect - connect things to channels

    def disconnect(self, channel_ids, thing_ids, token):
        DisconnectThing - disconnect thing and channel

class Channels:
    def __init__(self, url):
        self.url = url

    def create(self, channel, token):
        CreateChannel - creates new channel and generates UUID
    
    def construct_query(self, params):
        ConstructQuery - allows you to set query params: offset, limit, order, direction

    def create_bulk(self, channels, token):
        CreateChannels - creates bulk of channels and generates channels UUID

    def get(self, channel_id, token):
        Channel - gets channel by ID

    def get_all(self, token, query_params=None):
        Channels - gets all channels

    def get_by_thing(self, thing_id, params, token):
        GetByThing - Gets all channels to which a specific channel is connected to

    def update(self, channel_id, token, channel):
        UpdateChannel - update a channel

    def delete(self, chanID, token):
        DeleteChannel - removes channel

class Messages:
    def __init__(self, url):
        self.url = url

    def send(self, channel_id, msg, token):
        Send - Sends message via HTTP protocol

    def read(self, channel_id, token):
        Read - Reads messages from database for a given channel
