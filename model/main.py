import os
import json
import flask
from flask import Flask, request, jsonify
from transformers import AutoTokenizer, AutoModelForCausalLM

# Create Flask application
app = Flask(__name__)

# Define your model path
MODEL_NAME = "meta-llama/Llama-3.2-1B" # Adjust if needed

# Set up Hugging Face token 
HF_TOKEN = os.getenv("HUGGINGFACE_TOKEN")

# Global variables for model and tokenizer
tokenizer = None
model = None

def load_model():
    """Load the tokenizer and model."""
    global tokenizer, model
    try:
        # Load tokenizer with padding token handling
        tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME, use_auth_token=HF_TOKEN)
        
        # Add padding token if not already present
        if tokenizer.pad_token is None:
            tokenizer.pad_token = tokenizer.eos_token
        
        # Load model
        model = AutoModelForCausalLM.from_pretrained(MODEL_NAME, use_auth_token=HF_TOKEN)
        
        # Resize model embedding if you added a new token
        model.resize_token_embeddings(len(tokenizer))
        
        return True
    except Exception as e:
        print(f"Error loading model: {e}")
        return False

# Load predefined Q&A
def load_predefined_qa():
    """Load predefined Q&A from JSON file."""
    try:
        with open('qnatemplate.json') as f:
            return json.load(f)
    except FileNotFoundError:
        print("Error: 'qnatemplate.json' file not found.")
        return []

# Predefined Q&A list
COLLEGE_QA = load_predefined_qa()

def get_predefined_answer(query):
    """Check if the query matches a predefined Q&A."""
    for qa in COLLEGE_QA:
        if query.lower() in qa["question"].lower():
            return qa["answer"]
    return None

def get_model_response(query):
    """Fetch response from LLaMA model."""
    if tokenizer is None or model is None:
        return "Model is not loaded. Please check the setup."

    # Encode input with attention mask
    inputs = tokenizer(query, return_tensors="pt", padding=True, truncation=True)
    
    # Generate response
    outputs = model.generate(
        inputs.input_ids,
        attention_mask=inputs.attention_mask,
        max_length=50,  # Adjust length as needed
        pad_token_id=tokenizer.eos_token_id
    )
    
    # Decode and return response
    response = tokenizer.decode(outputs[0], skip_special_tokens=True)
    return response

@app.route('/query', methods=['POST'])
def handle_query():
    """Handle incoming query requests."""
    # Validate request
    if not request.is_json:
        return jsonify({"error": "Request must be JSON"}), 400
    
    # Extract query
    data = request.get_json()
    query = data.get('query')
    
    if not query:
        return jsonify({"error": "No query provided"}), 400
    
    # Try predefined answers first
    predefined_answer = get_predefined_answer(query)
    if predefined_answer:
        return jsonify({
            "response": predefined_answer,
            "source": "predefined"
        })
    
    # Fallback to model response
    model_response = get_model_response(query)
    return jsonify({
        "response": model_response,
        "source": "model"
    })

@app.route('/health', methods=['GET'])
def health_check():
    """Simple health check endpoint."""
    model_status = "loaded" if model is not None else "not loaded"
    return jsonify({
        "status": "healthy",
        "model_status": model_status
    })

def main():
    # Load the model before starting the server
    if not load_model():
        print("Failed to load model. Server may not function correctly.")
    
    # Run the Flask app
    app.run(host='0.0.0.0', port=5000, debug=True)

if __name__ == "__main__":
    main()