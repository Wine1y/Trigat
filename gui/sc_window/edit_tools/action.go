package editTools

type ToolAction interface {
	Undo()
	Redo()
}

type actionNode struct {
	action   ToolAction
	previous *actionNode
	next     *actionNode
}

type ActionsQueue struct {
	undoNode *actionNode
	redoNode *actionNode
}

func NewActionsQueue() *ActionsQueue {
	return &ActionsQueue{}
}

func (queue ActionsQueue) CanUndo() bool {
	return queue.undoNode != nil
}

func (queue ActionsQueue) CanRedo() bool {
	return queue.redoNode != nil
}

func (queue *ActionsQueue) Push(action ToolAction) {
	node := &actionNode{action: action}
	if queue.undoNode != nil {
		node.previous = queue.undoNode
		queue.undoNode.next = node
	}
	if queue.redoNode != nil {
		node.next = queue.redoNode
		queue.redoNode.previous = node
	}
	queue.undoNode = node
}

func (queue *ActionsQueue) Undo() {
	queue.undoNode.action.Undo()
	queue.redoNode = queue.undoNode
	queue.undoNode = queue.undoNode.previous
}

func (queue *ActionsQueue) Redo() {
	queue.redoNode.action.Redo()
	queue.undoNode = queue.redoNode
	queue.redoNode = queue.redoNode.next
}
