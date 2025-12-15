from behave import *
from hamcrest import *
import requests
import tempfile

@given(u'I have files with contents "{content_a}" and "{content_b}" to scan with v2')
def step_impl(context, content_a, content_b):
	context.file_content_a = tempfile.NamedTemporaryFile(delete=False)
	context.file_content_a.write(content_a.encode('utf-8'))
	context.file_content_a.close()

	context.file_content_b = tempfile.NamedTemporaryFile(delete=False)
	context.file_content_b.write(content_b.encode('utf-8'))
	context.file_content_b.close()

@when('I v2/scan the files for a virus')
def step_impl(context): 
	files = [('file', open(context.file_content_a.name, 'rb')), ('file', open(context.file_content_b.name, 'rb'))]
	url = context.clamrest
	r = requests.post(url + "/v2/scan", files=files)
	context.result = r

@then('I get a http status of "{status}" from v2/scan with multiple files')
def step_impl(context, status):
	assert_that(context.result.status_code, equal_to(int(status)))
