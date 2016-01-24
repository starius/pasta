local lxsh = require 'lxsh'
local config = require("lapis.config").get()

local highlight = {}

local function getHighlighter(filename)
    for _, lang in ipairs {'lua', 'c', 'bib', 'sh'} do
        if filename:match('%.' .. lang .. '$') then
            return lxsh.highlighters[lang]
        end
    end
end

function highlight.highlight(content, filename)
    if #content > config.max_highlighted then
        return nil, "Too long"
    end
    local highlighter = getHighlighter(filename)
    if not highlighter then
        return nil, "Unknown file type"
    end
    return highlighter(content, {
        formatter = lxsh.formatters.html,
    })
end

return highlight
