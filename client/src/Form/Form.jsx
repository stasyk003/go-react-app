import React, { useState } from 'react';
import './Form.css';

export const Form = ({ books, onUpdateBook, onDeleteBook, onCreateBook }) => {
  const [title, setTitle] = useState('');
  const [author, setAuthor] = useState('');
  const [showList, setShowList] = useState(false);
  const [selectedBook, setSelectedBook] = useState(null);

  const handleSubmit = (event) => {
    event.preventDefault();
    onCreateBook({
      title: title,
      author: author
    });
    setTitle('');
    setAuthor('');
  };

  const handleUpdateBook = () => {
    if (selectedBook) {
      onUpdateBook(selectedBook.id, selectedBook);
      setSelectedBook(null);
    }
  };

  const handleDeleteBook = () => {
    if (selectedBook) {
      onDeleteBook(selectedBook.id);
      setSelectedBook(null);
    }
  };

  const handleSelectBook = (book) => {
    setSelectedBook(book);
  };

  const handleCloseList = () => {
    setShowList(false);
  }

  return (
    <form className="form-container" onSubmit={handleSubmit}>
      <div className="form-group">
        <label className="form-label" htmlFor="title">Title:</label>
        <input
          className="form-input"
          type="text"
          id="title"
          value={title}
          onChange={(event) => setTitle(event.target.value)}
        />
      </div>
      <div className="form-group">
        <label className="form-label" htmlFor="author">Author:</label>
        <input
          className="form-input"
          type="text"
          id="author"
          value={author}
          onChange={(event) => setAuthor(event.target.value)}
        />
      </div>
      <button className="form-submit" type="submit">Create Book</button>
      <button className="list-button" onClick={() => setShowList(!showList)}>List</button>
      {showList && (
        <div className="book-list-container">
          <div className="book-list">
            {books.map((book) => (
              <div key={book.id} className={`book-list-item ${selectedBook && selectedBook.id === book.id ? 'selected' : ''}`} onClick={() => handleSelectBook(book)}>
                <div className="book-list-title">Title: {book.title}</div>
                <div className="book-list-author">Author: {book.author}</div>
                <div className="book-list-actions">
                  <button className="update-button" onClick={() => handleUpdateBook()}>Update</button>
                  <button className="delete-button" onClick={() => handleDeleteBook()}>Delete</button>
                </div>
              </div>
            ))}
          </div>
          <button className="close-button" onClick={() => handleCloseList()}>Close</button>
        </div>
      )}
    </form>
  );
};

