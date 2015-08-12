from behave import *
from hamcrest import *
import requests

@given('I have a file with contents "{contents}"')
def step_imp(context, contents):
	context.file_contents = contents 	

@when('I scan the file for a virus')
def step_impl(context): 
	files = { 'file': context.file_contents }
	url = "http://127.0.0.1:9000"
	r = requests.post(url + "/scan", files=files)
	context.result = r

@then('I get a http status of "{status}"')
def step_impl(context, status):
	assert_that(int(status), equal_to(context.result.status_code))
