package controller

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/session"

	"github.com/gin-gonic/gin"
)

// The Clients page: the account-centric view of who exists, as opposed to the
// Inbounds page's view of what serves them.
//
// READ ONLY. Every mutation the page performs goes back through the existing
// client routes on /panel/api/inbounds (addClient, updateClient,
// delClientByEmail, bulkUpdateClients) carrying an inboundIds set. That is
// deliberate: those paths already carry the reseller pricing, the ownership
// assertions over every id, the same-protocol refusal and the projection, and a
// second write path would be a second place for all of that to drift out of step.
//
// The one account state those routes cannot express - an account NO inbound serves,
// which has no settings blob to be spliced into and no protocol to be addressed by -
// has two routes of its own, saveAccount and delAccount. They live on the same
// /panel/api/inbounds group and behind the same permission bits for exactly the
// reason above: the guards belong together, not on either side of a page boundary.
type ClientsController struct {
	BaseController
}

func NewClientsController(g *gin.RouterGroup) *ClientsController {
	a := &ClientsController{}
	a.initRouter(g)
	return a
}

func (a *ClientsController) initRouter(g *gin.RouterGroup) {
	// accessInbounds, the same claim the Inbounds page takes: this shows the same
	// accounts from the other side, so a caller who may see one may see the other.
	// The rows are then narrowed per caller inside ListAccounts.
	g.GET("/list", a.list)
	g.GET("/assignable", a.assignable)
	g.GET("/history", a.history)
}

// history returns bandwidth analytics only after the same account visibility
// check as the Clients list. It never exposes destinations, DNS queries, or
// packet contents.
func (a *ClientsController) history(c *gin.Context) {
	email := c.Query("email")
	user := session.GetLoginUser(c)
	if email == "" || user == nil {
		jsonObj(c, nil, errors.New("invalid client email"))
		return
	}
	if !user.IsSuperAdmin {
		allowed := false
		var err error
		if user.IsReseller {
			var result *service.AccountListResult
			result, err = accountService.ListAccounts(user, 1, 1, email, "newest")
			allowed = err == nil && len(result.Rows) == 1 && strings.EqualFold(result.Rows[0].Email, email)
		} else {
			allowed, err = accessService.CanAccessClientEmail(email, user.Id)
		}
		if err != nil || !allowed {
			jsonObj(c, nil, errors.New("client not accessible"))
			return
		}
	}
	rangeKey := c.DefaultQuery("range", "24h")
	durations := map[string]time.Duration{"1h": time.Hour, "24h": 24 * time.Hour, "7d": 7 * 24 * time.Hour, "30d": 30 * 24 * time.Hour}
	duration, ok := durations[rangeKey]
	if !ok {
		jsonObj(c, nil, errors.New("invalid history range"))
		return
	}
	points, err := TrafficHistory(email, time.Now().Add(-duration), time.Now())
	if err != nil {
		jsonObj(c, nil, err)
		return
	}
	jsonObj(c, map[string]any{"email": email, "range": rangeKey, "points": points}, nil)
}

// list returns one page of accounts the caller may see.
//
// The scoping lives in the service and is the whole security surface here: the
// accounts and memberships tables are panel-wide, so an unscoped read would hand
// every admin every other admin's customers and every reseller every other
// seller's. Passing the login user rather than a flag keeps that decision in one
// place.
// The sort is a query parameter and not a body field because this is a GET, and it
// is validated in the service rather than here: an unknown key falls back to
// "newest" there, so there is nothing for this handler to reject. There is no
// direction parameter on purpose - each ordering the page offers already says which
// way it runs. The result echoes back the ordering that was actually applied.
func (a *ClientsController) list(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	result, err := accountService.ListAccounts(session.GetLoginUser(c), page, size,
		c.Query("search"), c.Query("sort"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonObj(c, result, nil)
}

// assignable lists the inbounds the caller may put an account on, for the page's
// inbound picker.
//
// Filtered by the same grant the write path enforces, so the picker cannot offer
// an inbound the save would then refuse: an operator ticking a box and getting
// "not found" back with no explanation is the worst version of this.
func (a *ClientsController) assignable(c *gin.Context) {
	rows, err := accountService.AssignableInboundsFor(session.GetLoginUser(c))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonObj(c, rows, nil)
}
