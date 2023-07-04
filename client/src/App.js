import './App.css';
import { Form } from './Form/Form';
import React, { useState, useEffect } from 'react';
import axios from 'axios';

function App() {
    const [books, setBooks] = useState([]);
    const [showList, setShowList] = useState(false);
    
    useEffect(() => {
        getBooks();
    }, []);
    
    const getBooks = async () => {
        const response = await axios.get('http://localhost:8000/books');
        setBooks(response.data);
    };
    
    const createBook = async (book) => {
        await axios.post('http://localhost:8000/books', book);
        getBooks();
    };
    
    const updateBook = async (id, book) => {
        await axios.put('http://localhost:8000/books/${id}', book);
        getBooks();
    };
    
    const deleteBook = async (id) => {
        await axios.delete('http://localhost:8000/books/${id}');
        getBooks();
    };
    
    const handleCreateBook = async (book) => {
        await createBook(book);
    };
    
    const handleUpdateBook = async (id, book) => {
        await updateBook(id, book);
    };
    
    const handleDeleteBook = async (id) => {
        await deleteBook(id);
    };
    
    const handleShowList = () => {
        setShowList(!showList);
    };
    
    return (
        <div>
            <Form books={books} onUpdateBook={handleUpdateBook} onDeleteBook={handleDeleteBook} onCreateBook={handleCreateBook} />
        </div>
    );
}

export default App;