document.querySelectorAll('.edit-btn').forEach(function(btn) {
    btn.addEventListener('click', function(e) {
        e.preventDefault();
        const container = btn.closest('.task-info');
        container.querySelector('.task-title').style.display = 'none';
        container.querySelector('.edit-title').style.display = 'inline-block';
        btn.style.display = 'none';
        container.querySelector('.save-btn').style.display = 'inline-block';
        container.querySelector('.cancel-btn').style.display = 'inline-block';
        container.querySelector('.edit-done').disabled = false; // Habilita el checkbox
    });
});

document.querySelectorAll('.cancel-btn').forEach(function(btn) {
    btn.addEventListener('click', function() {
        const container = btn.closest('.task-info');
        container.querySelector('.task-title').style.display = 'inline';
        container.querySelector('.edit-title').style.display = 'none';
        container.querySelector('.edit-btn').style.display = 'inline';
        container.querySelector('.save-btn').style.display = 'none';
        btn.style.display = 'none';
        container.querySelector('.edit-done').disabled = true; // Deshabilita el checkbox
    });
});

document.querySelectorAll('.save-btn').forEach(function(btn) {
    btn.addEventListener('click', function() {
        const container = btn.closest('.task-info');
        const id = container.getAttribute('data-id');
        const newTitle = container.querySelector('.edit-title').value;
        const done = container.querySelector('.edit-done').checked ? "on" : "";

        fetch('/update', {
            method: 'POST',
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            body: `id=${encodeURIComponent(id)}&title=${encodeURIComponent(newTitle)}&done=${encodeURIComponent(done)}`
        }).then(resp => {
            if (resp.ok) {
                container.querySelector('.task-title').textContent = newTitle;
                container.querySelector('.task-title').style.display = 'inline';
                container.querySelector('.edit-title').style.display = 'none';
                container.querySelector('.edit-btn').style.display = 'inline';
                btn.style.display = 'none';
                container.querySelector('.cancel-btn').style.display = 'none';
                container.querySelector('.edit-done').disabled = true;
                // Actualiza el estilo si está completada
                if (done === "on") {
                    container.querySelector('.task-title').classList.add('completed');
                } else {
                    container.querySelector('.task-title').classList.remove('completed');
                }
            } else {
                alert('Error actualizando la tarea');
            }
        });
    });
});