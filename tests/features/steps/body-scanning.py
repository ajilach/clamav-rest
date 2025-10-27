from behave import *
from hamcrest import *
import requests

@given('I have a body with contents "{contents}"')
def step_imp(context, contents):
	context.body = contents

@when('I POST the content to /scanHandlerBody to scan the content for a virus')
def step_impl(context): 
	#files = {'file': context.file_contents}
	body = context.body
	url = context.clamrest
	r = requests.post(url + "/scanHandlerBody", data=body)
	context.result = r

@then('I get a http status of "{status}" from /scanHandlerBody')
def step_impl(context, status):
	#print("status var type: ")
	#print(type(status))
	#print("status var content: "+status)
	#print("response status type: ")
	#print(type(context.result.status_code))
	#print("response status: "+str(context.result.status_code))
	assert_that(int(status), equal_to(context.result.status_code))
