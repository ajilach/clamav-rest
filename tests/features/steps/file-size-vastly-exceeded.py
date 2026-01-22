from behave import *
from hamcrest import *
import requests
import tempfile
import os

@given(u'I have a too large file to scan with v2/scan')
def step_impl(context):
    fd, path = tempfile.mkstemp()
    try: 
        with os.fdopen(fd, "wb") as f:
            f.write(b"x" * (20 * 1024 * 1024))  # 20 MB
    except Exception:
        os.close(fd)
        raise

    context.file_name = path
    context.temp_file = True

@when('I v2/scan a too large file')
def step_impl(context): 
	files = {'file': open(context.file_name, 'rb')}
	url = context.clamrest
	r = requests.post(url + "/v2/scan", files=files)
	context.result = r

@then('I get a http status of "{status}" from v2/scan with a too large file')
def step_impl(context, status):
	assert_that(context.result.status_code, equal_to(int(status)))
