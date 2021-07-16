from functions import sdk

import json


s = sdk.SDK()

thing = {"thing_name": "thing"}
thing_id = "123-456-789"
thing_id1 = "123-223-333"
channel_id = "654-654-654"
channel_id1 = "654-654-654"
token = "9a8b7c6d5e4f3g21"
url = "http://localhost"
params = None


def test_create_thing(requests_mock):
    requests_mock.register_uri("POST", url + "/things", headers={"location": "/things/" + thing_id}, status_code=201)
    r = s.things.create(thing, token)
    assert r.error.status == 0
    assert thing_id == r.value


def test_create_existing_thing(requests_mock):
    requests_mock.register_uri("POST", url + "/things", headers={"location": "/things/" + thing_id}, status_code=409)
    r = s.things.create(thing, token)
    assert r.error.status == 1
    assert r.error.message == "Entity already exist."


def test_create_bulk_things(requests_mock):
    requests_mock.register_uri("POST", url + "/things/bulk", json=[thing_id, thing_id1], headers={"location": "/things/" + thing_id}, status_code=201)
    r = s.things.create_bulk(thing_id, token)
    assert r.error.status == 0
    assert [thing_id, thing_id1] == r.value


def test_create_bulk_things_missing_token(requests_mock):
    requests_mock.register_uri("POST", url + "/things/bulk", json=[thing_id, thing_id1], headers={"location": "/things/" + thing_id}, status_code=401)
    r = s.things.create_bulk(thing_id, token)
    assert r.error.status == 1
    assert r.error.message == "Missing or invalid access token provided."


def test_get_thing(requests_mock):
    requests_mock.register_uri("GET", url + "/things/" + thing_id, json=thing, status_code=200)
    r = s.things.get(thing_id, token)
    assert r.error.status == 0
    assert thing == r.value


def test_get_thing_malformed_query(requests_mock):
    requests_mock.register_uri("GET", url + "/things/" + thing_id, json=thing, status_code=400)
    r = s.things.get(thing_id, token)
    assert r.error.status == 1
    assert r.error.message == "Failed due to malformed query parameters."


def test_get_all_things(requests_mock):
    requests_mock.register_uri("GET", url + "/things", json=[thing_id, thing_id1], status_code=200)
    r = s.things.get_all(token)
    assert r.error.status == 0
    assert [thing_id, thing_id1] == r.value


def test_get_all_thing_does_not_exist(requests_mock):
    requests_mock.register_uri("GET", url + "/things", json=[thing_id, thing_id1], status_code=404)
    r = s.things.get_all(token)
    assert r.error.status == 1
    assert r.error.message == "Thing does not exist."


def test_get_by_channel(requests_mock):
    requests_mock.register_uri("GET", url + "/channels/" + channel_id + "/things", json=channel_id, headers={"Authorization": "/channels/" + channel_id + "/things"}, status_code=200)
    r = s.things.get_by_channel(channel_id, params, token)
    assert r.error.status == 0
    assert channel_id == r.value


def test_get_by_channel_missing_token(requests_mock):
    requests_mock.register_uri("GET", url + "/channels/" + channel_id + "/things", json=channel_id, headers={"Authorization": "/channels/" + channel_id + "/things"}, status_code=401)
    r = s.things.get_by_channel(channel_id, params, token)
    assert r.error.status == 1
    assert r.error.message == "Missing or invalid access token provided."


def test_update_thing(requests_mock):
    requests_mock.register_uri("PUT", url + "/things/" + thing_id, json=json.dumps(thing), status_code=200)
    r = s.things.update(thing_id, token, thing)
    assert r.error.status == 0


def test_update_thing_bad_json(requests_mock):
    requests_mock.register_uri("PUT", url + "/things/" + thing_id, json=json.dumps(thing), status_code=400)
    r = s.things.update(thing_id, token, thing)
    assert r.error.status == 1
    assert r.error.message == "Failed due to malformed JSON."


def test_delete_thing(requests_mock):
    requests_mock.register_uri("DELETE", url + "/things/" + thing_id, status_code=204)
    r = s.things.delete(thing_id, token)
    assert r.error.status == 0


def test_delete_bad_thing_id(requests_mock):
    requests_mock.register_uri("DELETE", url + "/things/" + thing_id, status_code=400)
    r = s.things.delete(thing_id, token)
    assert r.error.status == 1
    assert r.error.message == "Failed due to malformed thing's ID."


def test_connect_thing(requests_mock):
    requests_mock.register_uri("POST", url + "/connect", json=[channel_id, thing_id], status_code=201)
    r = s.things.connect(channel_id, thing_id, token)
    assert r.error.status == 0
    assert [channel_id, thing_id] == r.value


def test_connect_non_existing_entity(requests_mock):
    requests_mock.register_uri("POST", url + "/connect", json=[channel_id, thing_id], status_code=404)
    r = s.things.connect(channel_id, thing_id, token)
    assert r.error.status == 1
    assert r.error.message == "A non-existent entity request."


def test_disconnect_thing(requests_mock):
    requests_mock.register_uri("DELETE", url + "/channels/" + channel_id + "/things/" + thing_id, status_code=204)
    r = s.things.disconnect(channel_id, thing_id, token)
    assert r.error.status == 0


def test_disconnect_thing_or_channel_does_not_exist(requests_mock):
    requests_mock.register_uri("DELETE", url + "/channels/" + channel_id + "/things/" + thing_id, status_code=404)
    r = s.things.disconnect(channel_id, thing_id, token)
    assert r.error.status == 1
    assert r.error.message == "Channel or thing does not exist."
