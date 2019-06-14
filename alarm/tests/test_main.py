# -*- coding: utf-8 -*-

# -- stdlib --

# -- third party --

# -- own --
from base import TestBase
from example_data import alarm_example


# -- code --


class TestMain(TestBase):

    def test_process_single_event(self):
        from main import process_single_event
        from httmock import response, HTTMock, urlmatch

        @urlmatch(netloc=r'(api.nexmo.com|yunpian.com)')
        def response_content(url, request):
            print(url)
            headers = {'Content-Type': 'application/json'}
            if url.netloc == 'api.nexmo.com':
                return response(200, '{}', headers)
            elif url.netloc == 'yunpian.com':
                return response(200, '{"code": 0}', headers)
            else:
                raise Exception('Meh!')

        with HTTMock(response_content):
            process_single_event(alarm_example)
