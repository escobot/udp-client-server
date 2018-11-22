"""
LA1
Comp 445 - Fall 2018

HTTP Client Application
"""
import socket
import argparse
import sys
import httplib


def parse_get(args):
    """Parses the arguments from the command line in order to make a GET request"""
    url = args.url
    verbose = args.verbose
    if args.headers:
        headers = dict(kv.split(":") for kv in args.headers)
    else:
        headers = None
    res = httplib.get(url, headers, verbose)


def parse_post(args):
    """Parses the arguments from the command line in order to make a GET request"""
    url = args.url
    verbose = args.verbose
    # Headers
    if args.headers:
        headers = dict(kv.split(":") for kv in args.headers)
    else:
        headers = None

    # file or data
    data = ""
    if args.inline_data:
        data = args.inline_data
    else:
        if args.file:
            data = args.file  # To do file handling
        else:
            print("POST request needs inline_data or file")

    res = httplib.post(url, data, headers, verbose)
    return


# Command parser setup
parser = argparse.ArgumentParser(
    prog="httpc",
    usage="httpc (get|post) URL [-v] (-h \"k:v\")* [-d inline-data] [-f file]",
    description="Httpc is a command-line tool that uses TCP sockets to support HTTP protocol operations",
    epilog="",
    add_help=False)

subparsers = parser.add_subparsers()

# Get parser setup
get_parser = subparsers.add_parser("get", description="Sends a GET request to the specified URL", add_help=False)

get_parser.add_argument("url", help="The URL of the targeted HTTP server", metavar="URL")
get_parser.add_argument("-v", help="Prints more information on the requests and responses",
                        action="store_true", dest="verbose")
get_parser.add_argument("-h", help="adds headers to the request. expected format: \"Key:value\".",
                        metavar="key:value", action="append", default=[], dest="headers")
get_parser.set_defaults(func=parse_get)

# Post parser setup
post_parser = subparsers.add_parser("post", description="Sends a POST request to the specified URL", add_help=False)
post_parser.add_argument("url", help="The URL of the targeted HTTP server", metavar="URL")
post_parser.add_argument("-v", help="Prints more information on the requests and responses",
                         action="store_true", dest="verbose")
post_parser.add_argument("-h", help="adds headers to the request. expected format: \"Key:value\".",
                         metavar="key:value", action="append", default=[], dest="headers")
data_group = post_parser.add_mutually_exclusive_group(required=True)
data_group.add_argument("-d", help="Adds data to the body of a HTTP POST request",
                        metavar="inline-data", dest="inline_data", default="")
data_group.add_argument("-f", help="Adds the content of a file to the body of a HTTP POST request.",
                        metavar="file", dest="file")
post_parser.set_defaults(func=parse_post)

# Parse arguments from command line
args = parser.parse_args()

# print(args)   # For Debug
try:
    args.func(args)
except AttributeError:
    # Prints the usage information if no command is given
    parser.print_usage()

"""
For demo purposes
run these

python httpc.py get http://httpbin.org/get

python httpc.py get httpbin.org -h "Host: httpbin.org" -h "Accept: text/html,application/json"

python httpc.py get http://httpbin.org/status/418

python httpc.py post http://httpbin.org/post -d "Nice teapot you got there."

python httpc.py get http://httpbin.org/redirect-to?url=http://httpbin.org/get"&"status_code=302 -v
"""

