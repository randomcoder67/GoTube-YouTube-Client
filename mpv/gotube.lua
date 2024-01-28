#!/usr/bin/env lua

function tprint (tbl, indent)
  if not indent then indent = 0 end
  local toprint = string.rep(" ", indent) .. "{\r\n"
  indent = indent + 2 
  for k, v in pairs(tbl) do
	toprint = toprint .. string.rep(" ", indent)
	if (type(k) == "number") then
	  toprint = toprint .. "[" .. k .. "] = "
	elseif (type(k) == "string") then
	  toprint = toprint  .. k ..  "= "   
	end
	if (type(v) == "number") then
	  toprint = toprint .. v .. ",\r\n"
	elseif (type(v) == "string") then
	  toprint = toprint .. "\"" .. v .. "\",\r\n"
	elseif (type(v) == "table") then
	  toprint = toprint .. tprint(v, indent + 2) .. ",\r\n"
	else
	  toprint = toprint .. "\"" .. tostring(v) .. "\",\r\n"
	end
  end
  toprint = toprint .. string.rep(" ", indent-2) .. "}"
  return toprint
end

function getChapters(f)
	table = {}
	i = 1
	for line in f:lines() do
		entry = {}
		j = 0
		for str in string.gmatch(line .. "DELIM", "(.-)(DELIM)") do
			entry[j] = str
			j = j + 1
		end
		if entry[0] == "" then
			break
		end
		table[i] = {time = tonumber(entry[1]), title = entry[0]}
		i = i + 1
	end
	return table
end

function addChapters()
	videoId = mainData[tonumber(mp.get_property("playlist-pos"))][2]
	handle = io.popen("gotube --fork --get-video-data " .. videoId)
	playtimeURL = handle:read()
	watchtimeURL = handle:read()
	
	markWatchedURLs[videoId] = {playtimeURL, watchtimeURL}
	--Print(tprint(markWatchedURLs))
	chapters = getChapters(handle)
	mp.set_property_native('chapter-list', chapters)
end

local options = {
	folderName = "",
	quality = ""
}

markWatchedURLs = {}

require "mp.options".read_options(options, "gotube")
--os.execute("notify-send " .. options.folderName)

function Print(thing)
	os.execute("notify-send \'" .. thing .. "\'")
end

function readPlaylistFile(folderName)
	--os.execute("notify-send \"FOLDER:" .. folderName .. "/details.txt\"")
	f = io.open(folderName .. "/details.txt", "r")
	content = f:read()
	data = {}
	
	i = 0
	while (content ~= nil) do
		entry = {}
		j = 0
		for str in string.gmatch(content .. "DELIM", "(.-)(DELIM)") do
			entry[j] = str
			j = j + 1
		end
		data[i] = entry
		i = i + 1
		content = f:read()
	end
	
	return data
end

mainData = readPlaylistFile(options.folderName)

function thing()
	--os.execute("sleep 2")
	--os.execute("notify-send \"thing: " .. tostring(mp.get_property("playlist-pos")) .. "\"")
	videoId = mainData[tonumber(mp.get_property("playlist-pos"))][2]
	handle = io.popen("gotube --fork --get-quality " .. videoId .. " " .. options.quality)
	directLink = handle:read()
	--os.execute("notify-send " .. directLink)
	handle:close()
	mp.set_property("stream-open-filename", directLink)
	--async os.execute("sleep 10")
	--[[
	curPos = mp.get_property("playlist-pos")
	playLen = mp.get_property("playlist-count")
	if (curPos ~= "100") then
		Print("HERE")
		handle = io.popen("gotube -getlink " .. stuff[tonumber(curPos)][1])
		mp.commandv("loadfile", result, "append")
		
		mp.commandv("playlist-move", tostring(playLen), tostring(curPos+1))
		mp.commandv("playlist-remove", tostring(curPos+2))
		
		Print("DONE")
	else
		Print("OTHER")
	end
	]]--
end

function watched()
	--Print("watched")
	videoId = mainData[tonumber(mp.get_property("playlist-playing-pos"))][2]
	time = mp.get_property("time-pos")
	--Print("Time: " .. time)
	
	--Print("In func: " .. videoId)
	--Print(tprint(markWatchedURLs))
	
	os.execute("gotube --fork --mark-watched " .. videoId .. " \"" .. tostring(time) .. "\" \"" .. markWatchedURLs[videoId][1] .. "\" \"" ..  markWatchedURLs[videoId][2] .. "\"")
end

mp.add_hook("on_load", 50, thing)
mp.add_hook("on_unload", 50, watched)
mp.register_event("file-loaded", addChapters)