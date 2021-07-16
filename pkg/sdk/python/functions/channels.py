import requests

from functions import response
from functions import errors


class Channels:
    def __init__(self, url):
        self.url = url

    def create(self, channel, token):
        '''Creates channel entity in the database'''
        mf_resp = response.Response()
        http_resp = requests.post(self.url + "/channels", json=channel, headers={"Authorization": token})
        if http_resp.status_code != 201:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.channels["create"], http_resp.status_code)
        else:
            location = http_resp.headers.get("location")
            mf_resp.value = location.split('/')[2]
        return mf_resp

    def create_bulk(self, channels, token):
        '''Creates multiple channels in a bulk'''
        mf_resp = response.Response()
        http_resp = requests.post(self.url + "/channels/bulk", json=channels, headers={"Authorization": token})
        if http_resp.status_code != 201:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.channels["create_bulk"], http_resp.status_code)
        else:
            mf_resp.value = http_resp.json()
        return mf_resp

    def get(self, chanID, token):
        '''Gets a channel entity for a logged-in user'''
        mf_resp = response.Response()
        http_resp = requests.get(self.url + "/channels/" + chanID, headers={"Authorization": token})
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.channels["get"], http_resp.status_code)
        else:
            mf_resp.value = http_resp.json()
        return mf_resp

    def construct_query(self, params):
        if params is None:
            return ""
        query = '?'
        param_types = ['offset', 'limit', 'order', 'direction']
        for pt in param_types:
            if pt in params.keys():
                query += "{}={}&".format(pt, params[pt])
        return query


    def get_all(self, token, query_params=None):
        '''Gets all channels from database'''
        query = self.construct_query(query_params)
        url = self.url + '/channels' + query
        mf_resp = response.Response()
        http_resp = requests.get(url, headers={"Authorization": token})
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.channels["get_all"], http_resp.status_code)
        else:
            mf_resp.value = http_resp.json()
        return mf_resp

    def get_by_thing(self, thing_id, params, token):
        '''Gets all channels to which a specific thing is connected to'''
        query = self.construct_query(params)
        url = self.url + "/things/" + thing_id + '/channels' + query
        mf_resp = response.Response()
        http_resp = requests.get(url, headers={"Authorization": token})
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.channels["get_by_thing"], http_resp.status_code)
        else:
            mf_resp.value = http_resp.json()
        return mf_resp

    def update(self, channel_id, token, channel):
        '''Updates channel entity'''
        http_resp = requests.put(self.url + "/channels/" + channel_id, json=channel, headers={"Authorization": token})
        mf_resp = response.Response()
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.channels["update"], http_resp.status_code)
        return mf_resp

    def delete(self, chanID, token):
        '''Deletes a channel entity from database'''
        http_resp = requests.delete(self.url + "/channels/" + chanID, headers={"Authorization": token})
        mf_resp = response.Response()
        if http_resp.status_code != 204:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.channels["delete"], http_resp.status_code)
        return mf_resp
