"""
LA1
Comp 445 - Fall 2018

HTTP Client Library
"""
import socket
import argparse
import sys
import json
from urllib.parse import urlparse


class Response:
    """Represents an HTTP response message.

    Attributes:
        http_version (str): HTTP version.
        code (int): Status code.
        status (str): Description of status code.
        headers (dict): Collection of key value pairs representing the response
            headers.
        body (str): The response body.
    """

    def __init__(self, response: str):
        """Parse the response string."""
        # The first consecutive CRLF sequence demarcates the start of the
        # entity-body.
        preamble, self.body = response.split("\r\n\r\n", maxsplit=1)
        status_line, *headers = preamble.split("\r\n")
        self.http_version, code, *status = status_line.split()
        self.code = int(code)
        self.status = " ".join(status)
        map(_remove_whitespace, headers)
        self.headers = dict(kv.split(":", maxsplit=1) for kv in headers)

    def __str__(self):
        """Return a string representation of the response."""
        status_line = "{} {} {}".format(self.http_version, self.code,
                                        self.status)
        headers = "\n".join(
            "{}: {}".format(k, v) for k, v in self.headers.items())
        return "\n".join((status_line, headers, self.body))


def _remove_whitespace(s: str):
    """Return a string with all whitespace removed from the input."""
    return "".join(s.split())


def handle_recv(sock: socket.socket):
    """
    Handles the receiving of a response from a server
    """
    BUFFER_SIZE = 8*1024
    response = b""
    while True:
        data = sock.recv(BUFFER_SIZE)
        response = response + data
        if not data:
            break

    return response


def send_req(url: str, port: int, req: str, verbose: bool):
    """
    Sends a request to the specified url:port
    and returns the response as a string
    :param url:
    :param port:
    :param req:
    :param verbose: enables verbose mode
    :return:
    """
    res = ""
    conn = socket.create_connection((url, port), 5)

    try:
        if verbose:
            print("Sending: \n" + req)  # print request

        data = req.encode("UTF-8")
        conn.sendall(data)

        res_data = handle_recv(conn)
        res_str = res_data.decode("UTF-8")
        res = Response(res_str)
        if verbose:
            print(res.headers)
        print(res.body)
    except socket.timeout:
        print("Connection to " + url + ":" + str(port) + " timed out")

    finally:
        conn.close()
        return res


def get(url: str, headers=None, verbose=False):
    """
    Makes a GET request and returns the response
    :param url:
    :param verbose: enables verbose mode
    :param headers:
    :param body:
    :return: response from server
    """

    # Parse URL components
    url_parsed = urlparse(url)
    # if scheme is absent from URL, urlparse makes problems
    # this next part checks if the url was parsed correctly
    if url_parsed.hostname is None:
        url = "http://"+url
        url_parsed = urlparse(url, "http")  # if not parsed properly, parse the url again
    host = url_parsed.hostname
    port = url_parsed.port or 80
    path = url_parsed.path or "/"
    query = url_parsed.query or ""
    uri = path
    if query != "":
        uri = uri + "?" + query

    # Prepare request line
    req = "GET " + uri + " HTTP/1.0\r\n"

    # Headers
    if headers is None:
        headers = {}
        headers.setdefault("Host", host)
        headers.setdefault("User-Agent", "HttpClient-Concordia")

    # Add headers to request
    for k, v in headers.items():
        req = req + k + ":" + v + "\r\n"
    # The request needs to finish with two empty lines. took me 4 hours to figure this out on my own
    req = req + "\r\n"

    # Send request
    res = send_req(host, port, req, verbose)
    # print("Reply: \n" + res)  # print response
    if res.code >= 300 and res.code < 400:
        print("Redirecting to: " + res.headers['Location'])
        url = res.headers['Location'].strip()
        return get(url, headers, verbose)

    return res


def post(url: str, data="", headers=None, verbose=False):
    """
    Sends a POST request
    :param url:
    :param data: the body of the request
    :param verbose: enables verbose mode
    :param headers:
    :return:
    """
    # Parse URL components
    url_parsed = urlparse(url)
    # if scheme is absent from URL, urlparse makes problems
    # this next part checks if the url was parsed correctly
    if url_parsed.hostname is None:
        url = "http://"+url
        url_parsed = urlparse(url, "http")  # if not parsed properly, parse the url again
    host = url_parsed.hostname
    port = url_parsed.port or 80
    path = url_parsed.path or "/"
    query = url_parsed.query or ""
    uri = path
    if query != "":
        uri = uri + "?" + query

    # Prepare request line
    req = "POST " + uri + " HTTP/1.0\r\n"

    # Headers
    if headers is None:
        headers = {}
        headers.setdefault("Host", host)
        headers.setdefault("User-Agent", "HttpClient-Concordia")
        headers.setdefault("Content-Length", str(len(data)))

    # Add headers to request
    for k, v in headers.items():
        req = req + k + ":" + v + "\r\n"
    req = req + "\r\n"
    req = req + data + "\r\n"
    # The request needs to finish with two empty lines. took me 4 hours to figure this out on my own
    req = req + "\r\n"

    # Send request
    res = send_req(host, port, req, verbose)
    # print("Reply: \n" + res)  # print response
    return res


"""
For Demo purposes
"""
"""
#get("localhost:8007", True)
get("http://httpbin.org/get", True)
get("http://httpbin.org/status/418", True)
post("http://httpbin.org/post", "Nice teapot you got there.", True)
"""
# get("http://httpbin.org/status/418", None, True)
# post("http://httpbin.org/post", "Nice teapot you got there.", None, True)
# get("https://httpbin.org/redirect-to?url=http://httpbin.org/get&status_code=302", None, False)
