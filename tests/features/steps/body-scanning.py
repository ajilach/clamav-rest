from behave import *
from hamcrest import *
import requests

@given('I have a body with contents "{contents}"')
def step_impl(context, contents):
	context.body = contents

@when('I POST the content to /scanHandlerBody to scan the content for a virus')
def step_impl(context): 
	body = context.body
	url = context.clamrest
	r = requests.post(url + "/scanHandlerBody", data=body)
	context.result = r

@then('I get a http status of "{status}" from /scanHandlerBody')
def step_impl(context, status):
	assert_that(context.result.status_code, equal_to(int(status)))
