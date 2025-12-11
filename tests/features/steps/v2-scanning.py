from behave import *
from hamcrest import *
import requests

@given('I have a file with contents "{contents}" to scan with v2')
def step_impl(context, contents):
	context.file_contents = contents

@when('I v2/scan the file for a virus')
def step_impl(context): 
	files = {'file': context.file_contents}
	url = context.clamrest
	r = requests.post(url + "/v2/scan", files=files)
	context.result = r

@then('I get a http status of "{status}" from v2/scan')
def step_impl(context, status):
	assert_that(context.result.status_code, equal_to(int(status)))
