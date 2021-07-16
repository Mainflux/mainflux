class Error():
    def __init__(self, status, message):
        self.status = status
        self.message = message



class Response(object):
    def __init__(self, value=None, error_status=0, error_message=""):
        self.error = Error(error_status, error_message)
        self.value = value
