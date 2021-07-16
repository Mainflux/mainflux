def handle_error(error_dict, status_code):
    if status_code in error_dict.keys():
        return error_dict[status_code]
    elif status_code in errors.keys():
        return errors[status_code]
    else:
        return "Unknown error"


errors = {
    400: "Failed due to malformed JSON.",
    401: "Missing or invalid access token provided.",
    403: "Missing or invalid access token provided.",
    404: "A non-existent entity request.",
    409: "Entity already exist.",
    415: "Missing or invalid content type.",
    422: "Database can't process request.",
    500: "Unexpected server-side error occurred.",
}
users = {
    "create": {
        409: "Failed due to using an existing email address.",
    },
    "login": {
        409: "Failed due to using an existing email address.",
    },
    "get": {
        400: "Failed due to malformed query parameters.",
    },
    "get_all": {
        400: "Failed due to malformed query parameters.",
    },
    "update": {
        404: "Failed due to non existing user.",
    },
    "update_password": {

    },
}

things = {
    "create": {
        422: "Unprocessable Entity."
    },
    "create_bulk": {

    },
    "get": {
        400: "Failed due to malformed query parameters.",
        404: "Thing does not exist.",
    },
    "get_all": {
        404: "Thing does not exist.",
    },
    "get_by_channel": {
        400: "Failed due to malformed query parameters.",
    },
    "update": {
        404: "Thing does not exist.",
    },
    "delete": {
        400: "Failed due to malformed thing's ID.",
    },
    "connect": {

    },
    "disconnect": {
        400: "Failed due to malformed query parameters.",
        404: "Channel or thing does not exist.",
    },
}

channels = {
    "create": {

    },
    "create_bulk": {

    },
    "get": {

    },
    "get_all": {
        400: "Failed due to malformed channel's ID.",
        404: "Channel does not exist.",
    },
    "get_by_thing": {
        400: "Failed due to malformed query parameters.",
        404: "Thing does not exist.",
    },
    "update": {
        404: "Channel does not exist."
    },
    "delete": {
        400: "Failed due to malformed channel's ID."
    },
}

messages = {
    "send": {
        400: "Message discarded due to its malformed content.",
        403: "Message discarded due to missing or invalid credentials.",
        404: "Message discarded due to invalid channel id.",
        415: "Message discarded due to invalid or missing content type.",
    },
    "read": {
        400: "Failed due to malformed query parameters.",
    },
}

groups = {
    "create": {
        409: "Failed due to using an existing email address.",
    },
    "get": {
        400: "Failed due to malformed query parameters.",
        404: "Group does not exist.",
    },
    "get_all": {
        400: "Failed due to malformed query parameters.",
        404: "Group does not exist.",
    },
    "update": {
        400: "Failed due to malformed query parameters.",
        404: "Group does not exist.",
    },
    "members": {
        409: "Failed due to using an existing email address.",
    },
    "assign": {

    },
    "unassign": {
        400: "Failed due to malformed query parameters.",
        404: "Group does not exist.",
    },
    "delete": {
        400: "Failed due to malformed query parameters.",
        404: "Group does not exist.",
    },
}
