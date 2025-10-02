from flask import Flask, request, jsonify
import re
import hmac
import hashlib
import os

app = Flask(__name__)
KEY = os.environ["APP_SECRET"]

# TODO: rate limit
@app.route("/login", methods=["POST"])
def login():
    data = request.get_json()
    code = data.get("code", "")

    if not re.fullmatch(r"\d{3}-\d{3}-\d{3}", code):
        return jsonify({"error": "Invalid code. Must be in the format XXX-XXX-XXX."}), 400

    if code == "123-456-789":
        token = hmac.new(KEY.encode(), code.encode(), hashlib.sha256).hexdigest()
        return jsonify({"message": "Login successful!", "token": token})
    else:
        return jsonify({"error": "Unauthorized"}), 401
    

if __name__ == "__main__":
    app.run(debug=False, host="0.0.0.0", port=80)
