from behave import *
from hamcrest import *
import requests

@given('I have a path with contents "{contents}" to scan with scanPath')
def step_imp(context, contents):
	context.query_param = contents

@when('I scan the path for a virus')
def step_impl(context): 
	#files = {'file': context.file_contents}
	url = context.clamrest
	r = requests.get(url + "?"+ context.query_param)
	context.result = r

@then('I get a http status of "{status}" from scanPath')
def step_impl(context, status):
    print("status: "+ status)
    print("context status: "+ str(context.result.status_code))
    assert_that(context.result.status_code, equal_to(int(status)))
