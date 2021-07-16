from functions import sdk


s = sdk.SDK()

msg = "helloMF"
token = "9a8b7c6d5e4f3g21"
url = "http://localhost"
channel_id = "224-335-668"


def test_send(requests_mock):
    requests_mock.register_uri("POST", url + "/http/channels/" + channel_id + "/messages", status_code=202)
    r = s.messages.send(channel_id, msg, token)
    assert r.error.status == 0


def test_read(requests_mock):
    requests_mock.register_uri("GET", url + "/channels/" + channel_id + "/messages", json=msg, status_code=200)
    r = s.messages.read(channel_id, token)
    assert r.error.status == 0
    assert msg == r.value
