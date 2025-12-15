from behave import *
from hamcrest import *
import requests

@given('I have a file path with contents "{contents}" to scan with scanFile')
def step_impl(context, contents):
	context.query_param = contents

@when('I scan the file path for a virus')
def step_impl(context): 
	baseUrl = context.clamrest
	url = baseUrl + "/scanFile?" + str(context.query_param)
	print(url)
	r = requests.get(url)
	context.result = r

@then('I get a http status of "{status}" from scanFile')
def step_impl(context, status):
	print("status: "+ status)
	print("context status: "+ str(context.result.status_code))
	assert_that(context.result.status_code, equal_to(int(status)))
