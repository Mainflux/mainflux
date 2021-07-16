import requests
import json

from functions import response
from functions import errors


class Users:
    def __init__(self, url):
        self.url = url

    def create(self, user):
        '''Creates user entity in the database'''
        mf_resp = response.Response()
        http_resp = requests.post(self.url + "/users", json=user)
        if http_resp.status_code != 201:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.users["create"], http_resp.status_code)
        else:
            location = http_resp.headers.get("location")
            mf_resp.value = location.split('/')[2]
        return mf_resp

    def login(self, user):
        '''Creates a user token'''
        mf_resp = response.Response()
        http_resp = requests.post(self.url + "/tokens", json=user)
        if http_resp.status_code != 201:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.users["login"], http_resp.status_code)
        else:
            mf_resp.value = http_resp.json()["token"]
        return mf_resp

    def get(self, user_id, token):
        '''Gets a user entity for a logged-in user'''
        mf_resp = response.Response()
        http_resp = requests.get(self.url + "/users/" + user_id, headers={"Authorization": token})
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.users["get"], http_resp.status_code)
        else:
            mf_resp.value = http_resp.json()
        return mf_resp

    def get_all(self, token):
        '''Gets all users from database'''
        http_resp = requests.get(self.url + "/users", headers={"Authorization": token})
        mf_resp = response.Response()
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.users["get_all"], http_resp.status_code)
        else:
            mf_resp.value = http_resp.json()
        return mf_resp

    def update(self, user, token):
        '''Updates user entity'''
        http_resp = requests.put(self.url + "/users", headers={"Authorization": token}, data=json.dumps(user))
        mf_resp = response.Response()
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.users["update"], http_resp.status_code)
        return mf_resp

    def update_password(self, old_password, password, token):
        '''Changes user password'''
        payload = {
          "old_password": old_password,
          "password": password
        }
        http_resp = requests.patch(self.url + "/password", headers={"Authorization": token}, json=payload)
        mf_resp = response.Response()
        if http_resp.status_code != 201:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.users["update_password"], http_resp.status_code)
        return mf_resp
