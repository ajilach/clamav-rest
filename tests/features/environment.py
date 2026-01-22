import os

def before_all(context):
	if os.environ.get('CLAMREST_HOST'):
		context.clamrest = "{}".format(os.environ.get('CLAMREST_HOST'))
	else:
		context.clamrest = "http://127.0.0.1:9000" 

def after_scenario(context, scenario):
    if hasattr(context, "file_name") and context.file_name:
        try:
            os.remove(context.file_name)
        except FileNotFoundError:
            pass
