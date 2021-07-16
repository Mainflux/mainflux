import requests

from functions import response
from functions import errors


class Groups:
    def __init__(self, url):
        self.url = url

    def create(self, group, token):
        '''Creates group entity in the database'''
        mf_resp = response.Response()
        http_resp = requests.post(self.url + "/groups", json=group, headers={"Authorization": token})
        if http_resp.status_code != 201:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.groups["create"], http_resp.status_code)
        else:
            location = http_resp.headers.get("location")
            mf_resp.value = location.split('/')[2]
        return mf_resp

    def get(self, group_id, token):
        '''Gets a group entity'''
        mf_resp = response.Response()
        http_resp = requests.get(self.url + "/groups/" + group_id, headers={"Authorization": token})
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.groups["get"], http_resp.status_code)
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

    def get_all(self, group_id, token, query_params=None):
        '''Gets all groups from database'''
        query = self.construct_query(query_params)
        url = self.url + '/groups/' + group_id + query
        mf_resp = response.Response()
        http_resp = requests.get(url, headers={"Authorization": token})
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.groups["get_all"], http_resp.status_code)
        else:
            mf_resp.value = http_resp.json()
        return mf_resp

    def update(self, group_id, token, group):
        '''Updates group entity'''
        http_resp = requests.put(self.url + "/groups/" + group_id, json=group, headers={"Authorization": token})
        mf_resp = response.Response()
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.groups["update"], http_resp.status_code)
        return mf_resp

    def members(self, group_id, token):
        '''Get list of members ID's from group'''
        http_resp = requests.post(self.url + "/groups/" + group_id + "/members", headers={"Authorization": token})
        mf_resp = response.Response()
        if http_resp.status_code != 204:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.groups["members"], http_resp.status_code)
        return mf_resp

    def assign(self, group_id, token, members):
        '''Assign'''
        mf_resp = response.Response()
        http_resp = requests.post(self.url + "/groups/" + group_id + "/members", headers={"Authorization": token}, json=members)
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.groups["assign"], http_resp.status_code)
        return mf_resp

    def unassign(self, group_id, token, members):
        '''Assign'''
        mf_resp = response.Response()
        http_resp = requests.delete(self.url + "/groups/" + group_id + "/members", headers={"Authorization": token}, json=members)
        if http_resp.status_code != 204:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.groups["unassign"], http_resp.status_code)
        return mf_resp

    def delete(self, group_id, token):
        '''Deletes a group entity from database'''
        http_resp = requests.delete(self.url + "/groups/" + group_id, headers={"Authorization": token})
        mf_resp = response.Response()
        if http_resp.status_code != 204:
            mf_resp.error.status = 1
            mf_resp.error.message = errors.handle_error(errors.groups["delete"], http_resp.status_code)
        return mf_resp
