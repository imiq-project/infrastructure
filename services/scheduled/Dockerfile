FROM python:3.13
WORKDIR /app
RUN pip install requests
COPY harbor.tracks .
COPY main.py .
CMD [ "python", "-u", "main.py" ]
