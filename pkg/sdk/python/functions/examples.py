from functions import users
from functions import things
from functions import messages
from functions import channels
from functions import groups
import requests
from functions import SDK


default_url = "http://localhost"

sdk = SDK()

'''To start working with the Mainflux system, you need to create a user account'''
mf_resp = sdk.users.create({"email": "<user_email>", "password": "<user_password>"})
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''To log in to the Mainflux system, you need to create a user token'''
mf_resp = sdk.users.login({"email": "<user_email>", "password": "<user_password>"})
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''You can always check the user entity that is logged in by entering the user ID and token'''
mf_resp = sdk.users.get(<user_id>, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Updating user entities in the database'''
mf_resp = sdk.users.update(<user_token>, {"metadata": {"foo": "bar"}})
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''You can get all users in the database by calling the get_all () function'''
mf_resp = sdk.users.get_all(<user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Changing the user password can be done by calling the update password function'''
mf_resp = sdk.users.update_password({"old_password":"<old_password>", "password":"<new_password>"}, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''To create a thing, you need the thing name and a user token'''
mf_resp = sdk.things.create({"name": "<thing_name>"}, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''You can create multiple things at once by entering a series of things structures and a user token'''
mf_resp = sdk.things.create_bulk({"name": "<thing_name_1>"}, {"name": "<thing_name_2>"}, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''You can get thing information by entering the thing ID and user token'''
mf_resp = sdk.things.get(<user_token>, <thing_id>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''You can get all things in the database by calling the get_all () function'''
mf_resp = sdk.things.get_all(<user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

"You can get all thing connected to channel"
mf_resp = sdk.things.get_by_channel(<channel_id>, {query_params}, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

"you can get all channels connected to thing"
mf_resp = sdk.channels.get_by_thing(<thing_id>, {query_params}, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Updating a thing entity in a database'''
mf_resp = sdk.things.update(<thing_id>, <user_token>, {"name": "<thing_name>"})
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''To delete a thing you need a thing ID and a user token'''
mf_resp = sdk.things.delete(<thing_id>, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Connect thing to channel'''
mf_resp = sdk.things.connect([<channel_id>], [<thing_id>], <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Disconnect thing from channel'''
mf_resp = sdk.things.disconnect([<channel_id>], [<thing_id>], <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''To create a channel, you need a channel and a token'''
mf_resp = sdk.channels.create({"name": "channel_name"}, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''As with things, you can create multiple channels at once'''
mf_resp = sdk.channels.create_bulk([{"name": "<channel_name_1>"}, {"name": "<channel_name_2>"}], <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Update channel entities in the database'''
mf_resp = sdk.channels.update(<channel_id>, <user_token>, {"name": "<channel_name_3>"})
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''A list of all the channels to which a given thing is connected'''
mf_resp = sdk.channels.get_by_thing(<thing_id>, {query_params}, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Delete channels from the database'''
mf_resp = sdk.channels.delete(<channel_id>, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''To create a group, you need the group name and a user token'''
mf_resp = sdk.groups.create({"name": "group_name"}, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''You can get group information by entering the group ID and token'''
mf_resp = sdk.groups.get(<group_id>, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Group update'''
mf_resp = sdk.groups.update(<group_id>, <user_token>,{<metadata>})
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''You can get groups in the database by calling the get_all () function'''
mf_resp = sdk.groups.get_all(<user_token>, <group_id>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Assign user, thing or channel to a group'''
mf_resp = sdk.groups.assign(<group_id>, <user_token>,{"members":["<user_id>" | "<thing_id_>" | "<channel_id_>"], "type":["users" | "things" | "channels"]})
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Unassign'''
mf_resp = sdk.groups.unassign(<group_id>, <user_token>,{"members":["<user_id>" | "<thing_id_>" | "<channel_id_>"], "type":["users" | "things" | "channels"]})
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Get list of members ID's from group'''
mf_resp = sdk.groups.get(<group_id>, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Delete group from the database'''
mf_resp = sdk.groups.delete(<group_id>, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Sends message via HTTP protocol'''
mf_resp = sdk.messages.send(<channel_id>, <msg>, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)

'''Reads messages from database for a given channel'''
mf_resp = sdk.messages.read(<channel_id>, <user_token>)
if mf_resp.error.status == 0:
    print(mf_resp.value)
else:
    print(mf_resp.error.message)
