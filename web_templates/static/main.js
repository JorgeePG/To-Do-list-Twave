document.querySelectorAll('.edit-btn').forEach(function (btn) {
    btn.addEventListener('click', function (e) {
        e.preventDefault();
        const container = btn.closest('.task-info');
        container.querySelector('.task-title').style.display = 'none';
        const editTitle = container.querySelector('.edit-title');
        editTitle.style.display = 'block'; // <-- Cambiado de 'flex' a 'block'
        editTitle.focus();
        btn.style.display = 'none';
        container.querySelector('.save-btn').style.display = 'inline-block';
        container.querySelector('.cancel-btn').style.display = 'inline-block';
        container.querySelector('.edit-done').disabled = false;
    });
});

document.querySelectorAll('.cancel-btn').forEach(function (btn) {
    btn.addEventListener('click', function () {
        const container = btn.closest('.task-info');
        container.querySelector('.task-title').style.display = 'inline';
        container.querySelector('.edit-title').style.display = 'none';
        container.querySelector('.edit-btn').style.display = 'inline-block';
        container.querySelector('.save-btn').style.display = 'none';
        btn.style.display = 'none';
        container.querySelector('.edit-done').disabled = true;
    });
});

document.querySelectorAll('.save-btn').forEach(function (btn) {
    btn.addEventListener('click', function () {
        const container = btn.closest('.task-info');
        const id = container.getAttribute('data-id');
        const newTitle = container.querySelector('.edit-title').value;
        const done = container.querySelector('.edit-done').checked ? "on" : "";

        fetch('/update', {
            method: 'POST',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body: `id=${encodeURIComponent(id)}&title=${encodeURIComponent(newTitle)}&done=${encodeURIComponent(done)}`
        }).then(resp => {
            if (resp.ok) {
                const taskTitle = container.querySelector('.task-title');
                taskTitle.textContent = newTitle;
                taskTitle.style.display = 'inline';
                container.querySelector('.edit-title').style.display = 'none';
                container.querySelector('.edit-btn').style.display = 'inline-block';
                btn.style.display = 'none';
                container.querySelector('.cancel-btn').style.display = 'none';
                container.querySelector('.edit-done').disabled = true;
                // Actualiza el estilo si está completada
                if (done === "on") {
                    taskTitle.classList.add('completed');
                } else {
                    taskTitle.classList.remove('completed');
                }
            } else {
                alert('Error actualizando la tarea');
            }
        });
    });
});