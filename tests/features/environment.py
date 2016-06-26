import os

def before_all(context):
	if os.environ.get('CLAMREST_HOST'):
		context.clamrest = "{}".format(os.environ.get('CLAMREST_HOST'))
	else:
		context.clamrest = "http://127.0.0.1:9000" 
