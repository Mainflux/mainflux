from functions import sdk

import json


s = sdk.SDK()

channel_id = "654-559-774"
members = {"members": channel_id, "type": "channels"}
group = {"group_name": "group"}
group_id = "888-888-888"
group_id1 = "989-787-686"
token = "9a8b7c6d5e4f3g21"
url = "http://localhost"


def test_create_group(requests_mock):
    requests_mock.register_uri("POST", url + "/groups", headers={"location": "/groups/" + group_id}, status_code=201)
    r = s.groups.create(group, token)
    assert r.error.status == 0
    assert group_id == r.value


def test_create_group_existing_email_address(requests_mock):
    requests_mock.register_uri("POST", url + "/groups", headers={"location": "/groups/" + group_id}, status_code=409)
    r = s.groups.create(group, token)
    assert r.error.status == 1
    assert r.error.message == "Failed due to using an existing email address."


def test_get_group(requests_mock):
    requests_mock.register_uri("GET", url + "/groups/" + group_id, json=group, status_code=200)
    r = s.groups.get(group_id, token)
    assert r.error.status == 0
    assert group == r.value


def test_get_group_does_not_exist(requests_mock):
    requests_mock.register_uri("GET", url + "/groups/" + group_id, json=group, status_code=404)
    r = s.groups.get(group_id, token)
    assert r.error.status == 1
    assert r.error.message == "Group does not exist."


def test_get_all_groups(requests_mock):
    requests_mock.register_uri("GET", url + "/groups/" + group_id, json=group_id, status_code=200)
    r = s.groups.get_all(group_id, token)
    assert r.error.status == 0
    assert group_id == r.value


def test_get_all_groups_malformed_query(requests_mock):
    requests_mock.register_uri("GET", url + "/groups/" + group_id, json=group_id, status_code=400)
    r = s.groups.get_all(group_id, token)
    assert r.error.status == 1
    assert r.error.message == "Failed due to malformed query parameters."


def test_update_group(requests_mock):
    requests_mock.register_uri("PUT", url + "/groups/" + group_id, json=json.dumps(group), status_code=200)
    r = s.groups.update(group_id, token, group)
    assert r.error.status == 0


def test_update_group_sverver_error(requests_mock):
    requests_mock.register_uri("PUT", url + "/groups/" + group_id, json=json.dumps(group), status_code=500)
    r = s.groups.update(group_id, token, group)
    assert r.error.status == 1
    assert r.error.message == "Unexpected server-side error occurred."


def test_members(requests_mock):
    requests_mock.register_uri("POST", url + "/groups/" + group_id + "/members", status_code=204)
    r = s.groups.members(group_id, token)
    assert r.error.status == 0


def test_members_bad_content_type(requests_mock):
    requests_mock.register_uri("POST", url + "/groups/" + group_id + "/members", status_code=415)
    r = s.groups.members(group_id, token)
    assert r.error.status == 1
    assert r.error.message == "Missing or invalid content type."


def test_assign(requests_mock):
    requests_mock.register_uri("POST", url + "/groups/" + group_id + "/members", status_code=200)
    r = s.groups.assign(group_id, token, members)
    assert r.error.status == 0


def test_assign_malformed_json(requests_mock):
    requests_mock.register_uri("POST", url + "/groups/" + group_id + "/members", status_code=400)
    r = s.groups.assign(group_id, token, members)
    assert r.error.status == 1
    assert r.error.message == "Failed due to malformed JSON."


def test_unassign(requests_mock):
    requests_mock.register_uri("DELETE", url + "/groups/" + group_id + "/members", status_code=204)
    r = s.groups.unassign(group_id, token, members)
    assert r.error.status == 0


def test_unassign_bad_token(requests_mock):
    requests_mock.register_uri("DELETE", url + "/groups/" + group_id + "/members", status_code=403)
    r = s.groups.unassign(group_id, token, members)
    assert r.error.status == 1
    assert r.error.message == "Missing or invalid access token provided."


def test_delete_group(requests_mock):
    requests_mock.register_uri("DELETE", url + "/groups/" + group_id, status_code=204)
    r = s.groups.delete(group_id, token)
    assert r.error.status == 0


def test_delete_group_does_not_exist(requests_mock):
    requests_mock.register_uri("DELETE", url + "/groups/" + group_id, status_code=404)
    r = s.groups.delete(group_id, token)
    assert r.error.status == 1
    assert r.error.message == "Group does not exist."
