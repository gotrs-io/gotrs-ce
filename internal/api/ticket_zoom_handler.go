package api

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/gotrs-io/gotrs-ce/internal/database"
    "github.com/gotrs-io/gotrs-ce/internal/repository"
)

// HandleTicketZoom renders an HTML ticket zoom view.
func HandleTicketZoom(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil || id <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid ticket id"})
        return
    }
    db, err := database.GetDB()
    if err != nil || db == nil {
        c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "ticket not found"})
        return
    }
    repo := repository.NewTicketRepository(db)
    t, err := repo.GetByID(uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "ticket not found"})
        return
    }
    r := GetPongo2Renderer()
    if r == nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "template renderer unavailable"})
        return
    }
    r.HTML(c, http.StatusOK, "pages/ticket_zoom.pongo2", gin.H{ "Ticket": t, "Title": t.Title })
}
