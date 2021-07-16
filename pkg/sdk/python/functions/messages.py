import requests

from functions import response
from functions import errors


class Messages:
    def __init__(self, url):
        self.url = url

    def send(self, channel_id, msg, token):
        '''Sends message via HTTP protocol'''
        url = self.url + "/http/channels/" + channel_id + '/messages'
        mf_resp = response.Response()
        http_resp = requests.post(url, json=msg, headers={"Authorization": token})
        if http_resp.status_code != 202:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.messages["send"], http_resp.status_code)
        return mf_resp

    def read(self, channel_id, token):
        '''Reads messages from database for a given channel'''
        url = self.url + "/channels/" + channel_id + '/messages'
        mf_resp = response.Response()
        http_resp = requests.get(url, headers={"Authorization": token})
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.messages["read"], http_resp.status_code)
        else:
            mf_resp.value = http_resp.json()
        return mf_resp
