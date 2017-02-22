from django.shortcuts import render
from django.http import HttpResponse

def files(request):
    return HttpResponse('Hi!')
