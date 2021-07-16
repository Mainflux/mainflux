from functions import sdk

import json

s = sdk.SDK()
user = {"email": "test@email.com", "password": "12345678"}
user1 = {"email": "test1@email.com", "password": "12345678"}
user2 = {"email": "test2@email.com", "password": "12345678"}
old_password = {"old_password": "12345678"}
password = {"password": "dsa"}
user_id = "123-456"
token = "9a8b7c6d5e4f3g21"
url = "http://localhost"


def test_create_user(requests_mock):
    requests_mock.register_uri("POST", url + "/users", headers={"location": "/users/" + user_id}, status_code=201)
    r = s.users.create(user)
    assert r.error.status == 0
    assert user_id == r.value


def test_login_user(requests_mock):
    requests_mock.register_uri("POST", url + "/tokens", json={"token": token}, status_code=201)
    r = s.users.login(user)
    assert r.error.status == 0
    assert token == r.value


def test_login_user_bad_email(requests_mock):
    requests_mock.register_uri("POST", url + "/tokens", status_code=409)
    r = s.users.login(user)
    assert r.error.status == 1
    assert r.error.message == "Failed due to using an existing email address."


def test_get_user(requests_mock):
    requests_mock.register_uri("GET", url + "/users/" + user_id, json=user, status_code=200)
    r = s.users.get(user_id, token)
    assert r.error.status == 0
    assert user == r.value


def test_get_user_bad_token(requests_mock):
    requests_mock.register_uri("GET", url + "/users/" + user_id, json=user, status_code=401)
    r = s.users.get(user_id, token)
    assert r.error.status == 1
    assert r.error.message == "Missing or invalid access token provided."


def test_get_all_users(requests_mock):
    requests_mock.register_uri("GET", url + "/users", json=[user, user2], status_code=200)
    r = s.users.get_all(token)
    assert r.error.status == 0
    assert [user, user2] == r.value


def test_get_all_user_bad_request(requests_mock):
    requests_mock.register_uri("GET", url + "/users/" + user_id, json=user, status_code=422)
    r = s.users.get(user_id, token)
    assert r.error.status == 1
    assert r.error.message == "Database can't process request."


def test_update_user(requests_mock):
    requests_mock.register_uri("PUT", url + "/users", json=json.dumps(user1), status_code=200)
    r = s.users.update(user, token)
    assert r.error.status == 0


def test_non_existing_user_update(requests_mock):
    requests_mock.register_uri("PUT", url + "/users", json=json.dumps(user1), status_code=404)
    r = s.users.update(user, token)
    assert r.error.status == 1
    assert r.error.message == "Failed due to non existing user."


def test_update_user_password(requests_mock):
    requests_mock.register_uri("PATCH", url + "/password", status_code=201)
    r = s.users.update_password(old_password, password, token)
    assert r.error.status == 0


def test_update_user_password_bad_token(requests_mock):
    requests_mock.register_uri("PATCH", url + "/password", status_code=415)
    r = s.users.update_password(old_password, password, token)
    assert r.error.status == 1
    assert r.error.message == "Missing or invalid content type."
