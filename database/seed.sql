-- Roles
INSERT INTO Roles (name) VALUES ('ADMIN');
INSERT INTO Roles (name) VALUES ('USER');
INSERT INTO Roles (name) VALUES ('MOD');

-- Users
INSERT INTO Users(username, email, password, created_at, role_id, avatar) VALUES ('admin', 'admin@gmail.com', '$2a$10$MGiiBL0svBk/QjdEu72lIeh/CycFFRKFVvyxrDqMKVy/GCsZ/z9Da', '2026-03-11 09:21:38.499123081+01:00', 1, 'avatars/Generic-Profile.jpg');
INSERT INTO Users(username, email, password, created_at, role_id, avatar) VALUES ('normal', 'normal@gmail.com','$2a$10$Ki.Nq6J64Ug4dgksLQsg6OmslX5yUfSAJDOeJmhIHflwDYl.XkjzC', '2026-03-11 09:27:50.168024652+01:00', 2,'avatars/Generic-Profile.jpg');

-- Categorie

INSERT INTO Categories(name,logo) VALUES ('Food', 'icons/Food.png');
INSERT INTO Categories(name,logo) VALUES ('Cinema', 'icons/Cinema.png');
INSERT INTO Categories(name,logo) VALUES ('Video games', 'icons/Video games.png');
INSERT INTO Categories(name,logo) VALUES ('Music', 'icons/Music.png');



-- Post
INSERT INTO Posts(user_id, title, content, created_at) VALUES (1, 'Title1', 'Content of the post n°1', '2026-03-08 09:21:38.499123081+01:00');
INSERT INTO Posts(user_id, title, content, created_at) VALUES (1, 'Title2', 'Content of the post n°2', '2026-03-09 09:21:38.499123081+01:00');
INSERT INTO Posts(user_id, title, content, created_at) VALUES (1, 'Title3', 'Content of the post n°3', '2026-03-10 09:21:38.499123081+01:00');
INSERT INTO Posts(user_id, title, content, created_at) VALUES (1, 'Title4', 'Content of the post n°4', '2026-03-11 09:21:38.499123081+01:00');
INSERT INTO Posts(user_id, title, content, created_at) VALUES (1, 'Title5', 'Content of the post n°5', '2024-02-22 09:21:38.499123081+01:00');



-- Comment
INSERT INTO Comments(user_id, post_id, content, created_at) VALUES (1, 1, 'Content of the Comment n°1', '2026-03-11 09:21:38.499123081+01:00');
INSERT INTO Comments(user_id, post_id, content, created_at) VALUES (1, 2, 'Content of the Comment n°2', '2026-03-11 09:21:38.499123081+01:00');
INSERT INTO Comments(user_id, post_id, content, created_at) VALUES (1, 3, 'Content of the Comment n°3', '2026-03-11 09:21:38.499123081+01:00');
INSERT INTO Comments(user_id, post_id, content, created_at) VALUES (2, 4, 'Content of the Comment n°4', '2026-03-11 09:21:38.499123081+01:00');


-- Likes
INSERT INTO Likes(user_id, post_id, comment_id, value) VALUES (1, 1, NULL, 1);
INSERT INTO Likes(user_id, post_id, comment_id, value) VALUES (2, 2, NULL, 1);
INSERT INTO Likes(user_id, post_id, comment_id, value) VALUES (1, 3, NULL, 1);
INSERT INTO Likes(user_id, post_id, comment_id, value) VALUES (2, 4, NULL, 1);


-- PostCategorie
INSERT INTO PostCategories (post_id, category_id) VALUES (1, 3);
INSERT INTO PostCategories (post_id, category_id) VALUES (2, 1);
INSERT INTO PostCategories (post_id, category_id) VALUES (3, 2);
INSERT INTO PostCategories (post_id, category_id) VALUES (4, 3);
INSERT INTO PostCategories (post_id, category_id) VALUES (5, 4);
